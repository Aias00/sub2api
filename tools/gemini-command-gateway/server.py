"""
Gemini Web Proxy - 本地 API 服务器

暴露 OpenAI 兼容的 REST API, 后端通过 Playwright 代理 gemini.google.com/app。

端点:
  POST /v1/chat/completions    - OpenAI 兼容接口 (支持流式)
  POST /chat                   - 简化接口
  GET  /health                 - 健康检查
  GET  /                       - API 文档

用法:
  python server.py             - 自动登录 (有已保存的会话则无头模式)
  python server.py --headed    - 强制有头模式
  python server.py --port 8080 - 指定端口
"""

import argparse
import asyncio
import json
import os
import re
import sys
import time
import uuid
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any, AsyncIterator, Optional

import uvicorn
from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, StreamingResponse
from pydantic import BaseModel

from gemini_client import GeminiClient

# ------------------------------------------------------------------ #
#  Account Pool Management                                            #
# ------------------------------------------------------------------ #

class AccountEntry:
    def __init__(self, session_dir: Path):
        self.session_dir = session_dir
        self.client = GeminiClient(session_dir=session_dir)
        self.lock = asyncio.Lock()
        self.id = session_dir.name

class AccountPool:
    def __init__(self, root_session_dir: Path = Path("./session")):
        self.root_dir = root_session_dir
        self.accounts: list[AccountEntry] = []
        self._index = 0

    async def discover(self):
        self.accounts = []
        if not self.root_dir.exists():
            self.root_dir.mkdir(parents=True, exist_ok=True)

        # Scan for subdirectories (each is an account)
        subdirs = []
        for d in self.root_dir.iterdir():
            if not d.is_dir() or d.name == "__pycache__":
                continue
            if not ((d / "state.json").exists() or (d / "cookies.json").exists()):
                continue
            subdirs.append(d)

        if not subdirs:
            # Fallback to root session for backward compatibility
            print(f"[AccountPool] No subdirectories found in {self.root_dir}, using root as single account.")
            self.accounts.append(AccountEntry(self.root_dir))
        else:
            for subdir in sorted(subdirs):
                self.accounts.append(AccountEntry(subdir))
            print(f"[AccountPool] Discovered {len(self.accounts)} accounts: {[a.id for a in self.accounts]}")

    def get_next(self) -> AccountEntry:
        if not self.accounts:
            raise RuntimeError("No accounts available in pool")
        entry = self.accounts[self._index]
        self._index = (self._index + 1) % len(self.accounts)
        return entry

    async def init_all(self, headless: bool = True):
        for acc in self.accounts:
            await acc.client.init(headless=headless)

    async def close_all(self):
        for acc in self.accounts:
            await acc.client.close()

# ------------------------------------------------------------------ #
#  全局状态                                                            #
# ------------------------------------------------------------------ #

_pool: Optional[AccountPool] = None
_pool_lock = asyncio.Lock()
_headless: bool = True
_bootstrap_chrome_user_data_dir: Optional[str] = None
_bootstrap_chrome_profile_directory: Optional[str] = None
_last_route_debug: dict[str, Any] = {}
_auth_sessions: dict[str, dict[str, Any]] = {}
_auth_sessions_lock = asyncio.Lock()

_DEFAULT_GEMINI_LOGIN_URL = "https://gemini.google.com/app"

_DEFAULT_ALLOWED_ORIGINS = [
    "http://localhost",
    "http://127.0.0.1",
    "http://localhost:3000",
    "http://127.0.0.1:3000",
    "http://localhost:5173",
    "http://127.0.0.1:5173",
]


def _env_bool(name: str, default: bool = False) -> bool:
    raw = os.getenv(name, "").strip().lower()
    if not raw:
        return default
    return raw in {"1", "true", "yes", "on"}


def _load_allowed_origins() -> list[str]:
    raw = os.getenv("GEMINI_PROXY_ALLOWED_ORIGINS", "")
    if not raw.strip():
        return _DEFAULT_ALLOWED_ORIGINS
    return [origin.strip() for origin in raw.split(",") if origin.strip()]


def _load_api_key() -> str:
    return os.getenv("GEMINI_PROXY_API_KEY", "").strip()


def _default_chrome_user_data_dir() -> str:
    env_override = os.getenv("GEMINI_PROXY_CHROME_USER_DATA_DIR", "").strip()
    if env_override:
        return os.path.expanduser(env_override)

    if sys.platform == "darwin":
        return os.path.expanduser("~/Library/Application Support/Google/Chrome")

    if sys.platform.startswith("linux"):
        return os.path.expanduser("~/.config/google-chrome")

    if sys.platform.startswith("win"):
        local_app_data = os.getenv("LOCALAPPDATA") or os.getenv("APPDATA") or ""
        if local_app_data:
            return os.path.join(local_app_data, "Google", "Chrome", "User Data")

    return ""


def _default_chrome_profile_directory() -> str:
    return os.getenv("GEMINI_PROXY_CHROME_PROFILE_DIRECTORY", "").strip() or "Default"


def _active_proxy_url() -> str:
    for name in (
        "GEMINI_PROXY_BROWSER_PROXY",
        "HTTPS_PROXY",
        "https_proxy",
        "HTTP_PROXY",
        "http_proxy",
        "ALL_PROXY",
        "all_proxy",
    ):
        value = os.getenv(name, "").strip()
        if value:
            return value
    return ""


_allowed_origins = _load_allowed_origins()
_api_key = _load_api_key()
_debug_requests = os.getenv("GEMINI_PROXY_DEBUG_REQUESTS", "").strip() == "1"
_local_auth_enabled = _env_bool("GEMINI_PROXY_LOCAL_AUTH_ENABLED", False)

# ------------------------------------------------------------------ #
#  模型名称映射                                                          #
#  Claude Code /model 切换 → Gemini 页面模型                             #
#                                                                      #
#  规则:                                                                 #
#    1. 包含 "haiku"  → Gemini 3 Flash                                  #
#    2. 其他全部       → Gemini 3.1 Pro                                  #
#                                                                      #
#  这样 Claude Code 中切到 haiku / haiku 4.5 时走 Flash，                #
#  sonnet / opus / 默认模型全部走 3.1 Pro。                               #
# ------------------------------------------------------------------ #

def resolve_gemini_model(model_name: str) -> str:
    """将任意 model 字符串解析为 Gemini 页面的模型关键词"""
    m = model_name.lower()

    # haiku / haiku 4.5 → Gemini 3 Flash (按钮菜单中的 "Fast")
    if "haiku" in m:
        return "Fast"

    # 其他全部 → 3.1 Pro (按钮菜单中的 "Pro")
    return "Pro"


@asynccontextmanager
async def lifespan(app: FastAPI):
    """
    lifespan 负责初始化账号池并启动所有客户端。
    """
    global _pool
    _pool = AccountPool()
    await _reload_account_pool()
    yield
    await _close_account_pool()


app = FastAPI(
    title="Gemini Web Proxy API",
    description="将 gemini.google.com/app 代理为本地 OpenAI 兼容 API",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=_allowed_origins,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ------------------------------------------------------------------ #
#  请求/响应模型                                                         #
# ------------------------------------------------------------------ #

class Message(BaseModel):
    role: str = "user"
    content: str


class ChatCompletionRequest(BaseModel):
    model: str = "gemini-web"
    messages: list[Message]
    stream: bool = False
    temperature: Optional[float] = None
    max_tokens: Optional[int] = None


class SimpleChatRequest(BaseModel):
    message: str
    conversation_id: str = ""
    stream: bool = False


class GeminiAuthStartRequest(BaseModel):
    login_id: Optional[str] = None
    account_id: Optional[int] = None
    login_mode: Optional[str] = None


class GeminiImportCookiesRequest(BaseModel):
    login_id: Optional[str] = None
    cookies_json: Optional[str] = None
    cookies: Optional[Any] = None


def _session_root_dir() -> Path:
    root = Path("./session")
    root.mkdir(parents=True, exist_ok=True)
    return root


def _normalize_login_id(login_id: Optional[str]) -> str:
    raw = (login_id or "").strip()
    if not raw:
        return "default"
    normalized = re.sub(r"[^a-zA-Z0-9._-]+", "-", raw).strip(".-")
    return normalized or "default"


def _session_dir_for_login(login_id: Optional[str]) -> Path:
    login_key = _normalize_login_id(login_id)
    return _session_root_dir() / login_key


def _existing_pool_account(login_id: Optional[str]) -> Optional[AccountEntry]:
    key = _normalize_login_id(login_id)
    if not _pool:
        return None
    for account in _pool.accounts:
        if account.id == key:
            return account
    return None


async def _close_account_pool():
    global _pool
    if _pool is None:
        return
    try:
        await _pool.close_all()
    except Exception as exc:
        print(f"[Gemini Proxy] 关闭账号池时出现警告: {exc}")


async def _reload_account_pool(headless: Optional[bool] = None):
    global _pool
    async with _pool_lock:
        if _pool is None:
            _pool = AccountPool()
        else:
            await _close_account_pool()
        await _pool.discover()
        if not _has_any_session_material():
            _pool.accounts = []
            return
        if _pool.accounts:
            await _pool.init_all(
                headless=_headless if headless is None else headless,
            )


def _has_any_session_material() -> bool:
    root = _session_root_dir()
    for subdir in root.iterdir():
        if not subdir.is_dir() or subdir.name == "__pycache__":
            continue
        if (subdir / "state.json").exists() or (subdir / "cookies.json").exists():
            return True
    return (root / "state.json").exists() or (root / "cookies.json").exists()


def _utc_now_rfc3339() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


async def _get_auth_session_payload(login_id: Optional[str]) -> Optional[dict[str, Any]]:
    key = _normalize_login_id(login_id)
    async with _auth_sessions_lock:
        session = _auth_sessions.get(key)
        if not session:
            return None
        payload = {k: v for k, v in session.items() if k != "task"}
        task = session.get("task")
        if task is not None and not task.done():
            payload["status"] = "pending"
        return payload


async def _set_auth_session_state(
    login_id: Optional[str],
    status: str,
    message: str,
    **extra: Any,
):
    key = _normalize_login_id(login_id)
    async with _auth_sessions_lock:
        session = _auth_sessions.get(key, {})
        task = session.get("task")
        session.update(
            {
                "login_id": key,
                "status": status,
                "message": message,
                "login_url": _DEFAULT_GEMINI_LOGIN_URL,
                "updated_at": _utc_now_rfc3339(),
                **extra,
            }
        )
        if task is not None:
            session["task"] = task
        _auth_sessions[key] = session


async def _run_local_auth_flow(login_id: str, session_dir: Path):
    client = GeminiClient(session_dir=session_dir)
    try:
        await client.init(
            headless=False,
            chrome_user_data_dir=_bootstrap_chrome_user_data_dir,
            chrome_profile_directory=_bootstrap_chrome_profile_directory,
        )
        # 登录完成后，账号池统一使用无头模式加载，避免再次弹出可见浏览器窗口。
        await _reload_account_pool(headless=True)
        await _set_auth_session_state(
            login_id,
            "ready",
            "Local browser login completed. Session saved.",
        )
    except Exception as exc:
        await _set_auth_session_state(
            login_id,
            "error",
            f"Local browser login failed: {exc}",
        )
    finally:
        try:
            await client.close()
        except Exception:
            pass


async def _launch_local_auth_flow(login_id: Optional[str]) -> dict[str, Any]:
    key = _normalize_login_id(login_id)
    session_dir = _session_dir_for_login(login_id)
    session_dir.mkdir(parents=True, exist_ok=True)

    async with _auth_sessions_lock:
        existing = _auth_sessions.get(key)
        if existing is not None:
            task = existing.get("task")
            if task is not None and not task.done():
                return {k: v for k, v in existing.items() if k != "task"}

        task = asyncio.create_task(_run_local_auth_flow(key, session_dir))
        _auth_sessions[key] = {
            "login_id": key,
            "status": "pending",
            "message": "Local browser launched. Complete sign-in in the Gemini window.",
            "login_url": _DEFAULT_GEMINI_LOGIN_URL,
            "updated_at": _utc_now_rfc3339(),
            "task": task,
        }

    return await _get_auth_session_payload(key) or {}


async def _session_status_payload(login_id: Optional[str]) -> dict[str, Any]:
    key = _normalize_login_id(login_id)
    session_dir = _session_dir_for_login(login_id)
    state_file = session_dir / "state.json"
    import_file = session_dir / "cookies.json"
    template_file = session_dir / "http_request_template.json"
    account = _existing_pool_account(login_id)

    status = "waiting_import"
    message = "Paste cookies JSON exported from your browser to activate Gemini Web."
    auth_session = await _get_auth_session_payload(login_id)
    if auth_session:
        status = str(auth_session.get("status") or status)
        message = str(auth_session.get("message") or message)
    if account and account.client._initialized:
        status = "ready"
        message = "Gemini Web session is ready."
    elif state_file.exists():
        status = "pending"
        message = "Session state detected. Gateway is loading it."
    elif import_file.exists():
        status = "pending"
        message = "Cookies imported. Gateway is preparing the session."
    elif _local_auth_enabled and not auth_session:
        message = "Start login to launch a local browser window, or import cookies as a fallback."

    return {
        "login_id": key,
        "status": status,
        "message": message,
        "login_url": auth_session.get("login_url") if auth_session else _DEFAULT_GEMINI_LOGIN_URL,
        "login_mode": auth_session.get("login_mode") if auth_session else "auto",
        "updated_at": auth_session.get("updated_at") if auth_session else _utc_now_rfc3339(),
        "expires_at": None,
        "has_state": state_file.exists(),
        "has_import": import_file.exists(),
        "has_template": template_file.exists(),
        "session_dir": str(session_dir),
        "local_auth_enabled": _local_auth_enabled,
    }


@app.post("/auth/start")
async def auth_start(payload: GeminiAuthStartRequest, request: Request):
    _require_api_key(request)
    session_dir = _session_dir_for_login(payload.login_id)
    session_dir.mkdir(parents=True, exist_ok=True)
    login_mode = (payload.login_mode or "").strip().lower() or "auto"
    if login_mode not in {"auto", "remote", "local"}:
        raise HTTPException(status_code=400, detail="invalid login_mode")

    if login_mode == "local" and not _local_auth_enabled:
        await _set_auth_session_state(
            payload.login_id,
            "waiting_import",
            "Local browser login is disabled on this gateway. Open Gemini in your current browser and import cookies instead.",
            login_mode=login_mode,
        )
    elif _local_auth_enabled and login_mode != "remote":
        await _launch_local_auth_flow(payload.login_id)
        await _set_auth_session_state(
            payload.login_id,
            "pending",
            "Local browser launched. Complete sign-in in the Gemini window.",
            login_mode=login_mode,
        )
    else:
        await _set_auth_session_state(
            payload.login_id,
            "waiting_import",
            "Open Gemini in your current browser, complete sign-in, then import cookies.",
            login_mode=login_mode,
        )
    return await _session_status_payload(payload.login_id)


@app.get("/auth/status")
async def auth_status(request: Request, login_id: Optional[str] = None):
    _require_api_key(request)
    return await _session_status_payload(login_id)


@app.post("/auth/import-cookies")
async def auth_import_cookies(payload: GeminiImportCookiesRequest, request: Request):
    _require_api_key(request)

    cookies_json = (payload.cookies_json or "").strip()
    if not cookies_json and payload.cookies is not None:
        cookies_json = json.dumps(payload.cookies, ensure_ascii=False)
    if not cookies_json:
        raise HTTPException(status_code=400, detail="cookies_json or cookies is required")

    try:
        parsed = json.loads(cookies_json)
    except json.JSONDecodeError as exc:
        raise HTTPException(status_code=400, detail=f"invalid cookies_json: {exc}") from exc

    session_dir = _session_dir_for_login(payload.login_id)
    session_dir.mkdir(parents=True, exist_ok=True)
    import_file = session_dir / "cookies.json"
    state_file = session_dir / "state.json"
    template_file = session_dir / "http_request_template.json"

    import_file.write_text(
        json.dumps(parsed, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    if state_file.exists():
        state_file.unlink()
    if template_file.exists():
        template_file.unlink()

    try:
        await _reload_account_pool()
    except Exception as exc:
        status = await _session_status_payload(payload.login_id)
        status["status"] = "pending"
        status["message"] = f"Cookies saved, but gateway initialization is still pending: {exc}"
        return JSONResponse(status_code=200, content=status)

    await _set_auth_session_state(
        payload.login_id,
        "ready",
        "Cookies imported. Session saved.",
    )
    status = await _session_status_payload(payload.login_id)
    if status["status"] != "ready":
        status["message"] = "Cookies saved. If initialization still shows pending, refresh status in a few seconds."
    return status


# Anthropic /v1/messages 格式
class AnthropicContentBlock(BaseModel):
    type: str = "text"
    text: str = ""


class AnthropicMessage(BaseModel):
    role: str
    content: str | list  # str 或 content blocks


class AnthropicRequest(BaseModel):
    model: str = "claude-3-5-sonnet-20241022"
    messages: list[AnthropicMessage]
    max_tokens: int = 4096
    stream: bool = False
    system: Optional[str | list] = None  # 兼容字符串和 content blocks 数组
    temperature: Optional[float] = None
    tools: Optional[list[dict[str, Any]]] = None
    tool_choice: Optional[dict[str, Any] | str] = None
    metadata: Optional[dict[str, Any]] = None
    thinking: Optional[dict[str, Any]] = None
    context_management: Optional[dict[str, Any]] = None
    output_config: Optional[dict[str, Any]] = None

    class Config:
        extra = "allow"


# ------------------------------------------------------------------ #
#  路由                                                                  #
# ------------------------------------------------------------------ #

@app.get("/")
async def root():
    return {
        "name": "Gemini Web Proxy API",
        "version": "1.0.0",
        "auth_required": bool(_api_key),
        "allowed_origins": _allowed_origins,
        "notes": [
            "This proxy is designed for single-user local use.",
            "Streaming responses are SSE-wrapped final chunks, not token-by-token output.",
        ],
        "endpoints": {
            "anthropic": "POST /v1/messages  (Claude Code 兼容)",
            "openai_compat": "POST /v1/chat/completions",
            "simple": "POST /chat",
            "health": "GET /health",
        },
        "docs": "/docs",
    }


@app.get("/health")
async def health():
    return {
        "status": "ok",
        "pool_ready": _pool is not None and len(_pool.accounts) > 0,
        "account_count": len(_pool.accounts) if _pool else 0,
        "accounts": [a.id for a in _pool.accounts] if _pool else [],
        "auth_required": bool(_api_key),
        "local_auth_enabled": _local_auth_enabled,
        "allowed_origins": _allowed_origins,
        "model_routing": {
            "haiku / haiku 4.5 → ": "Fast (Gemini 3 Flash)",
            "sonnet / opus / default / others → ": "Pro (Gemini 3.1 Pro)",
        },
    }


@app.get("/debug/last-route")
async def debug_last_route(request: Request):
    _require_api_key(request)
    return {
        "last_route": _last_route_debug or None,
    }


def _require_api_key(request: Request):
    if not _api_key:
        return

    auth_header = request.headers.get("authorization", "").strip()
    bearer_token = auth_header[7:].strip() if auth_header.lower().startswith("bearer ") else ""
    x_api_key = request.headers.get("x-api-key", "").strip()

    if bearer_token == _api_key or x_api_key == _api_key:
        return

    raise HTTPException(status_code=401, detail="缺少或无效的 API key")


def _maybe_warn_ignored_params(temperature: Optional[float], max_tokens: Optional[int]) -> list[str]:
    warnings = []
    if temperature is not None:
        warnings.append("temperature is currently ignored")
    if max_tokens is not None:
        warnings.append("max_tokens is currently ignored")
    return warnings


@app.get("/debug/screenshot")
async def debug_screenshot(request: Request):
    """截图 + 列出页面内所有按钮文字, 用于调试模型切换选择器 (使用第一个账号)"""
    _require_api_key(request)
    if not _pool or not _pool.accounts:
        raise HTTPException(status_code=503, detail="客户端未就绪")
    client = _pool.accounts[0].client
    if not client._initialized or not client._page:
        raise HTTPException(status_code=503, detail="账号 1 未就绪")
    page = client._page
    shot_path = "/tmp/gemini_debug.png"
    await page.screenshot(path=shot_path, full_page=False)
    buttons = await page.evaluate("""() => {
        const all = [...document.querySelectorAll('button, [role="button"], [role="option"], [role="menuitem"], mat-option')];
        return all.map(el => ({
            tag: el.tagName,
            text: el.innerText?.trim().slice(0, 80),
            ariaLabel: el.getAttribute('aria-label'),
            class: el.className?.toString().slice(0, 60),
        })).filter(el => el.text || el.ariaLabel);
    }""")
    return {
        "screenshot": shot_path,
        "button_count": len(buttons),
        "buttons": buttons[:60],
    }


@app.get("/debug/after-switch/{target}")
async def debug_after_switch(target: str, request: Request):
    """点击 mode picker → 点击目标模型 → 截图, 看切换后页面状态 (使用第一个账号)"""
    _require_api_key(request)
    if not _pool or not _pool.accounts:
        raise HTTPException(status_code=503, detail="客户端未就绪")
    client = _pool.accounts[0].client
    if not client._initialized or not client._page:
        raise HTTPException(status_code=503, detail="账号 1 未就绪")
    import asyncio as _asyncio
    page = client._page

    # 1. 点 mode picker
    picker = page.locator('button[aria-label="Open mode picker"]').first
    await picker.click()
    await _asyncio.sleep(1)

    # 2. 点目标模型 (JS 精确匹配标题)
    clicked = await page.evaluate(f"""() => {{
        const kw = '{target}'.toLowerCase();
        const btns = [...document.querySelectorAll('button.bard-mode-list-button')];
        for (const btn of btns) {{
            const title = (btn.innerText || '').split('\\n')[0].trim().toLowerCase();
            if (title === kw) {{ btn.click(); return btn.innerText.trim().slice(0,40); }}
        }}
        return null;
    }}""")
    await _asyncio.sleep(1.5)

    # 3. 截图
    shot = "/tmp/gemini_after_switch.png"
    await page.screenshot(path=shot)

    # 4. 看 mode picker 按钮现在显示什么
    picker_text = await page.locator('button[aria-label="Open mode picker"]').first.inner_text()

    # 5. 看页面上有没有弹窗/overlay
    overlay = await page.evaluate("""() => {
        const overlays = [...document.querySelectorAll('.cdk-overlay-container *')];
        return overlays.map(e => e.innerText?.trim()).filter(t=>t).slice(0,10);
    }""")

    return {
        "clicked": clicked,
        "picker_now_shows": picker_text.strip(),
        "screenshot": shot,
        "overlay_text": overlay,
    }


@app.get("/debug/mode-picker")
async def debug_mode_picker(request: Request):
    """点击 mode picker 按钮, 截图并列出菜单项 (使用第一个账号)"""
    _require_api_key(request)
    if not _pool or not _pool.accounts:
        raise HTTPException(status_code=503, detail="客户端未就绪")
    client = _pool.accounts[0].client
    if not client._initialized or not client._page:
        raise HTTPException(status_code=503, detail="账号 1 未就绪")
    import asyncio as _asyncio
    page = client._page

    # 点击 mode picker
    picker = page.locator('button[aria-label="Open mode picker"]').first
    visible = await picker.is_visible(timeout=3000)
    if not visible:
        return {"error": "Open mode picker button not found"}

    await picker.click()
    await _asyncio.sleep(1)

    shot_path = "/tmp/gemini_menu.png"
    await page.screenshot(path=shot_path, full_page=False)

    # 列出菜单里出现的所有可交互元素
    items = await page.evaluate("""() => {
        const all = [...document.querySelectorAll(
            '[role="menuitem"],[role="option"],[role="menu"] *,mat-option,.mat-menu-item,.cdk-overlay-container *'
        )];
        return all.map(el => ({
            tag: el.tagName,
            role: el.getAttribute('role'),
            text: el.innerText?.trim().slice(0, 100),
            ariaLabel: el.getAttribute('aria-label'),
            class: el.className?.toString().slice(0, 80),
        })).filter(el => el.text || el.ariaLabel);
    }""")

    # 关闭菜单
    await page.keyboard.press("Escape")

    return {
        "screenshot": shot_path,
        "picker_visible": visible,
        "menu_item_count": len(items),
        "menu_items": items[:50],
    }


@app.post("/v1/chat/completions")
async def openai_chat_completions(req: ChatCompletionRequest, request: Request):
    """OpenAI 兼容的 /v1/chat/completions 接口"""
    _require_api_key(request)
    if not _pool:
        raise HTTPException(status_code=503, detail="Gemini 账号池未就绪")

    prompt = _build_prompt(req.messages)
    gemini_model = resolve_gemini_model(req.model)

    acc = _pool.get_next()
    _record_route_debug("/v1/chat/completions", req.model, gemini_model, acc.id)
    warnings = _maybe_warn_ignored_params(req.temperature, req.max_tokens)

    if req.stream:
        return StreamingResponse(
            _openai_stream_generator(prompt, req.model, acc, gemini_model),
            media_type="text/event-stream",
            headers=_gemini_route_headers(gemini_model, acc.id),
        )
    else:
        async with acc.lock:
            text = await acc.client.chat(prompt, gemini_model=gemini_model)
        return JSONResponse(
            _openai_response(text, req.model, warnings=warnings),
            headers=_gemini_route_headers(gemini_model, acc.id),
        )


@app.post("/v1/messages")
async def anthropic_messages(req: AnthropicRequest, request: Request):
    """Anthropic /v1/messages 接口 — Claude Code CLI 兼容"""
    _require_api_key(request)
    if _debug_requests:
        try:
            body = await request.json()
            with open("/tmp/gemini_proxy_last_messages_request.json", "w", encoding="utf-8") as fh:
                json.dump(body, fh, ensure_ascii=False, indent=2)
        except Exception:
            pass
    if not _pool:
        raise HTTPException(status_code=503, detail="Gemini 账号池未就绪")

    gemini_model = resolve_gemini_model(req.model)
    acc = _pool.get_next()
    _record_route_debug("/v1/messages", req.model, gemini_model, acc.id)

    if req.stream:
        return StreamingResponse(
            _anthropic_stream_generator(req, acc, gemini_model),
            media_type="text/event-stream",
            headers=_gemini_route_headers(gemini_model, acc.id),
        )
    else:
        async with acc.lock:
            assistant_turn = await _generate_anthropic_assistant_turn(req, acc.client, gemini_model)
        return JSONResponse(
            _anthropic_response(
                assistant_turn,
                req.model,
                warnings=_maybe_warn_ignored_params(req.temperature, req.max_tokens),
            ),
            headers=_gemini_route_headers(gemini_model, acc.id),
        )


@app.post("/chat")
async def simple_chat(req: SimpleChatRequest, request: Request):
    """简化聊天接口"""
    _require_api_key(request)
    if not _pool:
        raise HTTPException(status_code=503, detail="Gemini 账号池未就绪")

    acc = _pool.get_next()
    _record_route_debug("/chat", "gemini-web", "Pro", acc.id)

    if req.stream:
        return StreamingResponse(
            _simple_stream_generator(req.message, req.conversation_id, acc),
            media_type="text/event-stream",
            headers=_gemini_route_headers("Pro", acc.id),
        )
    else:
        async with acc.lock:
            text = await acc.client.chat(req.message, req.conversation_id)
        return JSONResponse(
            {"reply": text, "model": "gemini-web"},
            headers=_gemini_route_headers("Pro", acc.id),
        )


# ------------------------------------------------------------------ #
#  流式生成器                                                            #
# ------------------------------------------------------------------ #

async def _openai_stream_generator(prompt: str, model: str, acc: AccountEntry, gemini_model: str = "2.5 Pro") -> AsyncIterator[bytes]:
    """生成 OpenAI 格式的 SSE 流"""
    completion_id = f"chatcmpl-{uuid.uuid4().hex[:12]}"
    created = int(time.time())

    async with acc.lock:
        # 发送 role delta
        yield _sse_chunk(
            completion_id, created, model,
            delta={"role": "assistant", "content": ""},
            finish_reason=None,
        )

        async for chunk in acc.client.stream_chat(prompt, gemini_model=gemini_model):
            yield _sse_chunk(
                completion_id, created, model,
                delta={"content": chunk},
                finish_reason=None,
            )

        # 结束
        yield _sse_chunk(
            completion_id, created, model,
            delta={},
            finish_reason="stop",
        )
    yield b"data: [DONE]\n\n"


async def _anthropic_stream_generator(req: AnthropicRequest, acc: AccountEntry, gemini_model: str = "2.5 Pro") -> AsyncIterator[bytes]:
    """生成 Anthropic SSE 格式的流 (Claude Code 所需格式)"""
    msg_id = f"msg_{uuid.uuid4().hex[:20]}"

    def sse(event: str, data: dict) -> bytes:
        return f"event: {event}\ndata: {json.dumps(data, ensure_ascii=False)}\n\n".encode()

    async with acc.lock:
        # message_start
        yield sse("message_start", {
            "type": "message_start",
            "message": {
                "id": msg_id, "type": "message", "role": "assistant",
                "content": [], "model": req.model,
                "stop_reason": None, "stop_sequence": None,
                "usage": {"input_tokens": 0, "output_tokens": 0},
            },
        })
        yield sse("ping", {"type": "ping"})

        queue: asyncio.Queue[tuple[str, object]] = asyncio.Queue()

        async def produce():
            try:
                assistant_turn = await _generate_anthropic_assistant_turn(req, acc.client, gemini_model)
                await queue.put(("turn", assistant_turn))
            except Exception as exc:
                await queue.put(("error", exc))

        producer = asyncio.create_task(produce())
        try:
            while True:
                try:
                    kind, payload = await asyncio.wait_for(queue.get(), timeout=5.0)
                except asyncio.TimeoutError:
                    yield sse("ping", {"type": "ping"})
                    continue
                if kind == "error":
                    raise payload
                assistant_turn = payload
                break
        finally:
            await producer

        output_tokens = assistant_turn["usage"]["output_tokens"]
        for index, block in enumerate(assistant_turn["content"]):
            block_type = block.get("type")
            if block_type == "text":
                yield sse("content_block_start", {
                    "type": "content_block_start",
                    "index": index,
                    "content_block": {"type": "text", "text": ""},
                })
                yield sse("content_block_delta", {
                    "type": "content_block_delta",
                    "index": index,
                    "delta": {"type": "text_delta", "text": block.get("text", "")},
                })
                yield sse("content_block_stop", {"type": "content_block_stop", "index": index})
                continue

            if block_type == "tool_use":
                yield sse("content_block_start", {
                    "type": "content_block_start",
                    "index": index,
                    "content_block": {
                        "type": "tool_use",
                        "id": block["id"],
                        "name": block["name"],
                        "input": {},
                    },
                })
                yield sse("content_block_delta", {
                    "type": "content_block_delta",
                    "index": index,
                    "delta": {
                        "type": "input_json_delta",
                        "partial_json": json.dumps(block.get("input", {}), ensure_ascii=False),
                    },
                })
                yield sse("content_block_stop", {"type": "content_block_stop", "index": index})

        yield sse("message_delta", {
            "type": "message_delta",
            "delta": {
                "stop_reason": assistant_turn["stop_reason"],
                "stop_sequence": None,
            },
            "usage": {"output_tokens": output_tokens},
        })

        # message_stop
        yield sse("message_stop", {"type": "message_stop"})


async def _simple_stream_generator(message: str, conv_id: str, acc: AccountEntry) -> AsyncIterator[bytes]:
    """简单 SSE 流"""
    async with acc.lock:
        async for chunk in acc.client.stream_chat(message, conv_id):
            data = json.dumps({"chunk": chunk}, ensure_ascii=False)
            yield f"data: {data}\n\n".encode()
    yield b"data: [DONE]\n\n"


# ------------------------------------------------------------------ #
#  辅助函数                                                              #
# ------------------------------------------------------------------ #

def _build_anthropic_prompt(req: AnthropicRequest) -> str:
    """将 Anthropic 格式请求转为单个 prompt 字符串"""
    parts = []
    if req.system:
        # system 可能是字符串或 content blocks 数组
        system_text = req.system
        if isinstance(system_text, list):
            system_text = " ".join(
                b.get("text", "") for b in system_text
                if isinstance(b, dict) and b.get("type") == "text"
            )
        if system_text:
            parts.append(f"System: {system_text}")
    for msg in req.messages:
        content = msg.content
        if isinstance(content, list):
            # content blocks: [{"type": "text", "text": "..."}]
            content = " ".join(
                b.get("text", "") for b in content if isinstance(b, dict) and b.get("type") == "text"
            )
        role = "User" if msg.role == "user" else "Assistant"
        parts.append(f"{role}: {content}")
    # 如果只有一条用户消息, 直接返回内容
    if len(req.messages) == 1 and not req.system:
        c = req.messages[0].content
        if isinstance(c, list):
            return " ".join(b.get("text", "") for b in c if isinstance(b, dict))
        return c
    return "\n".join(parts)


def _gemini_route_headers(gemini_model: str, account_id: str = "default") -> dict[str, str]:
    routed = "Fast" if gemini_model == "Fast" else "Pro"
    display = "Gemini 3 Flash" if routed == "Fast" else "Gemini 3.1 Pro"
    return {
        "X-Gemini-Routed-Model": routed,
        "X-Gemini-Routed-Display": display,
        "X-Gemini-Routed-Account": account_id,
    }


def _record_route_debug(endpoint: str, requested_model: str, gemini_model: str, account_id: str = "default"):
    global _last_route_debug
    routed = "Fast" if gemini_model == "Fast" else "Pro"
    _last_route_debug = {
        "endpoint": endpoint,
        "requested_model": requested_model,
        "routed_model": routed,
        "routed_display": "Gemini 3 Flash" if routed == "Fast" else "Gemini 3.1 Pro",
        "routed_account": account_id,
        "timestamp": int(time.time()),
    }


def _anthropic_response(assistant_turn: dict[str, Any], model: str, warnings: Optional[list[str]] = None) -> dict:
    payload = {
        "id": f"msg_{uuid.uuid4().hex[:20]}",
        "type": "message",
        "role": "assistant",
        "content": assistant_turn["content"],
        "model": model,
        "stop_reason": assistant_turn["stop_reason"],
        "stop_sequence": None,
        "usage": assistant_turn["usage"],
    }
    if warnings:
        payload["warnings"] = warnings
    return payload


async def _generate_anthropic_assistant_turn(req: AnthropicRequest, client: GeminiClient, gemini_model: str) -> dict[str, Any]:
    if _should_use_tool_bridge(req):
        prompt = _build_tool_bridge_prompt(req)
        raw = await client.chat(prompt, gemini_model=gemini_model)
        return _parse_tool_bridge_response(raw, req.tools or [])

    prompt = _build_anthropic_prompt(req)
    text = await client.chat(prompt, gemini_model=gemini_model)
    return _text_assistant_turn(text)


def _should_use_tool_bridge(req: AnthropicRequest) -> bool:
    if req.tools:
        return True
    for message in req.messages:
        if isinstance(message.content, list):
            for block in message.content:
                if isinstance(block, dict) and block.get("type") != "text":
                    return True
    return False


def _text_assistant_turn(text: str) -> dict[str, Any]:
    return {
        "content": [{"type": "text", "text": text}],
        "stop_reason": "end_turn",
        "usage": {"input_tokens": 0, "output_tokens": len(text) // 4},
    }


def _build_tool_bridge_prompt(req: AnthropicRequest) -> str:
    tools = [
        {
            "name": tool.get("name"),
            "description": tool.get("description", ""),
            "input_schema": tool.get("input_schema", {}),
        }
        for tool in (req.tools or [])
        if isinstance(tool, dict) and tool.get("name")
    ]

    rules = [
        "You are the model behind an Anthropic Messages API request used by Claude Code.",
        "You DO have access to the listed tools through the API. Do not claim that you cannot read files, inspect the repo, or write files when tools for that are available.",
        "Return ONLY valid JSON with no markdown fences and no prose outside JSON.",
        'Valid response shape: {"text":"optional user-facing text","tool_uses":[{"name":"ExactToolName","input":{}}]}',
        "Use exact tool names from the available tools list.",
        "The input for each tool must be a JSON object that matches the tool schema.",
        "If tools are needed to inspect the codebase, gather context, edit files, or run commands, prefer tool_uses over giving a speculative text answer.",
        "If the request is a repository task such as /init, code analysis, editing files, or running commands, start with tool calls instead of a final answer.",
        "When a previous user message contains tool results, use them to decide the next tool call or final answer.",
        "When you have enough information and no tool is needed, return text and an empty tool_uses array.",
        "Keep text concise. If you are making tool calls, text is usually empty or a very short status sentence.",
    ]

    if _tool_choice_requires_tool(req.tool_choice):
        rules.append("This request requires at least one tool call. Do not return an empty tool_uses array.")

    transcript_parts = []
    if req.system:
        transcript_parts.append("SYSTEM:\n" + _render_anthropic_content(req.system))
    for index, message in enumerate(req.messages, start=1):
        transcript_parts.append(
            f"{message.role.upper()} MESSAGE {index}:\n{_render_anthropic_content(message.content)}"
        )

    return "\n\n".join([
        "\n".join(rules),
        "AVAILABLE TOOLS:\n" + json.dumps(tools, ensure_ascii=False, indent=2),
        "CONVERSATION:\n" + "\n\n".join(transcript_parts),
        'NOW RETURN JSON ONLY. Example: {"text":"","tool_uses":[{"name":"Glob","input":{"pattern":"**/*.py"}}]}',
    ])


def _tool_choice_requires_tool(tool_choice: Any) -> bool:
    if tool_choice == "any":
        return True
    if isinstance(tool_choice, dict) and tool_choice.get("type") in {"any", "tool"}:
        return True
    return False


def _render_anthropic_content(content: Any) -> str:
    if isinstance(content, str):
        return content
    if not isinstance(content, list):
        return json.dumps(content, ensure_ascii=False)

    rendered = []
    for block in content:
        if not isinstance(block, dict):
            rendered.append(str(block))
            continue

        block_type = block.get("type")
        if block_type == "text":
            rendered.append(block.get("text", ""))
            continue
        if block_type == "tool_use":
            rendered.append(
                "<tool_use "
                f'name="{block.get("name", "")}" '
                f'id="{block.get("id", "")}">\n'
                f'{json.dumps(block.get("input", {}), ensure_ascii=False)}\n'
                "</tool_use>"
            )
            continue
        if block_type == "tool_result":
            rendered.append(
                "<tool_result "
                f'tool_use_id="{block.get("tool_use_id", "")}" '
                f'is_error="{bool(block.get("is_error"))}">\n'
                f'{_render_tool_result_content(block.get("content"))}\n'
                "</tool_result>"
            )
            continue
        rendered.append(json.dumps(block, ensure_ascii=False))

    return "\n".join(part for part in rendered if part)


def _render_tool_result_content(content: Any) -> str:
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for item in content:
            if isinstance(item, dict) and item.get("type") == "text":
                parts.append(item.get("text", ""))
            else:
                parts.append(json.dumps(item, ensure_ascii=False))
        return "\n".join(parts)
    return json.dumps(content, ensure_ascii=False)


def _parse_tool_bridge_response(raw: str, tools: list[dict[str, Any]]) -> dict[str, Any]:
    parsed = _extract_json_object(raw)
    if not isinstance(parsed, dict):
        return _text_assistant_turn(raw)

    allowed_names = {tool.get("name") for tool in tools if isinstance(tool, dict)}
    text = parsed.get("text", "")
    if not isinstance(text, str):
        text = json.dumps(text, ensure_ascii=False)

    tool_uses = parsed.get("tool_uses", [])
    if not isinstance(tool_uses, list):
        tool_uses = []

    content: list[dict[str, Any]] = []
    if text.strip():
        content.append({"type": "text", "text": text.strip()})

    for item in tool_uses:
        if not isinstance(item, dict):
            continue
        name = item.get("name")
        tool_input = item.get("input", {})
        if name not in allowed_names or not isinstance(tool_input, dict):
            continue
        content.append({
            "type": "tool_use",
            "id": f"toolu_{uuid.uuid4().hex[:20]}",
            "name": name,
            "input": tool_input,
        })

    if not content:
        return _text_assistant_turn(raw)

    output_tokens = max(len(text) // 4, 1)
    for block in content:
        if block["type"] == "tool_use":
            output_tokens += max(len(json.dumps(block["input"], ensure_ascii=False)) // 4, 1)

    stop_reason = "tool_use" if any(block["type"] == "tool_use" for block in content) else "end_turn"
    return {
        "content": content,
        "stop_reason": stop_reason,
        "usage": {"input_tokens": 0, "output_tokens": output_tokens},
    }


def _extract_json_object(raw: str) -> Any:
    cleaned = raw.strip()
    fenced = re.search(r"```(?:json)?\s*(\{.*\})\s*```", cleaned, re.DOTALL)
    if fenced:
        cleaned = fenced.group(1)
    else:
        start = cleaned.find("{")
        end = cleaned.rfind("}")
        if start != -1 and end != -1 and end > start:
            cleaned = cleaned[start:end + 1]
    try:
        return json.loads(cleaned)
    except Exception:
        return None


def _build_prompt(messages: list[Message]) -> str:
    """将 OpenAI 格式的消息列表转为单个 prompt"""
    if len(messages) == 1:
        return messages[-1].content

    parts = []
    for msg in messages:
        role = "用户" if msg.role == "user" else "助手" if msg.role == "assistant" else "系统"
        parts.append(f"{role}: {msg.content}")
    return "\n".join(parts)


def _openai_response(text: str, model: str, warnings: Optional[list[str]] = None) -> dict:
    payload = {
        "id": f"chatcmpl-{uuid.uuid4().hex[:12]}",
        "object": "chat.completion",
        "created": int(time.time()),
        "model": model,
        "choices": [
            {
                "index": 0,
                "message": {"role": "assistant", "content": text},
                "finish_reason": "stop",
            }
        ],
        "usage": {
            "prompt_tokens": -1,
            "completion_tokens": -1,
            "total_tokens": -1,
        },
    }
    if warnings:
        payload["warnings"] = warnings
    return payload


def _sse_chunk(
    completion_id: str,
    created: int,
    model: str,
    delta: dict,
    finish_reason: Optional[str],
) -> bytes:
    payload = {
        "id": completion_id,
        "object": "chat.completion.chunk",
        "created": created,
        "model": model,
        "choices": [
            {
                "index": 0,
                "delta": delta,
                "finish_reason": finish_reason,
            }
        ],
    }
    return f"data: {json.dumps(payload, ensure_ascii=False)}\n\n".encode()


# ------------------------------------------------------------------ #
#  登录预检                                                              #
# ------------------------------------------------------------------ #

async def _ensure_login():
    """
    在启动 uvicorn 前检查账号池中的登录状态。
    没有会话时仅输出提示，不阻塞服务启动，以便上层系统通过登录接口完成导入。
    """
    from pathlib import Path
    root_session = Path("./session")
    root_session.mkdir(exist_ok=True)

    # 如果没有任何子文件夹，检查根目录是否有 session
    subdirs = [d for d in root_session.iterdir() if d.is_dir() and d.name != "__pycache__"]
    if not subdirs:
        cookies_file = root_session / "state.json"
        import_file = root_session / "cookies.json"
        if cookies_file.exists() or import_file.exists():
            print("[Gemini Proxy] 检测到传统模式会话")
            return

        print("\n[Gemini Proxy] 当前没有任何已导入会话。")
        print("[Gemini Proxy] 你可以稍后通过 /auth/import-cookies 导入，或手动运行:")
        print("python scripts/import_cookies.py --dir user1")
        return

    print(f"[Gemini Proxy] 检测到 {len(subdirs)} 个账号会话")


# ------------------------------------------------------------------ #
#  入口                                                                  #
# ------------------------------------------------------------------ #

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Gemini Web Proxy API Server")
    parser.add_argument("--port", type=int, default=8000, help="监听端口 (默认 8000)")
    parser.add_argument("--host", default="127.0.0.1", help="监听地址 (默认 127.0.0.1)")
    parser.add_argument("--headed", action="store_true", help="强制使用有头浏览器")
    parser.add_argument(
        "--use-chrome-profile",
        action="store_true",
        help="首次登录时尝试复用本机 Google Chrome profile",
    )
    parser.add_argument(
        "--chrome-user-data-dir",
        default=_default_chrome_user_data_dir(),
        help="Google Chrome 用户数据目录 (也可用 GEMINI_PROXY_CHROME_USER_DATA_DIR)",
    )
    parser.add_argument(
        "--chrome-profile-directory",
        default=_default_chrome_profile_directory(),
        help="Google Chrome profile 目录名, 如 Default 或 Profile 1 (也可用 GEMINI_PROXY_CHROME_PROFILE_DIRECTORY)",
    )
    args = parser.parse_args()

    _headless = not args.headed
    _bootstrap_chrome_user_data_dir = (
        args.chrome_user_data_dir if args.use_chrome_profile else None
    )
    _bootstrap_chrome_profile_directory = (
        args.chrome_profile_directory if args.use_chrome_profile else None
    )

    print(f"\n[Gemini Proxy] 启动服务器 http://{args.host}:{args.port}")
    print(f"[Gemini Proxy] API 文档: http://{args.host}:{args.port}/docs")
    print(f"[Gemini Proxy] 浏览器模式: {'有头' if not _headless else '无头'}\n")
    if _active_proxy_url():
        print(f"[Gemini Proxy] 出站代理: {_active_proxy_url()}")
    if args.use_chrome_profile:
        print(
            "[Gemini Proxy] 首次登录将尝试复用 Chrome profile: "
            f"{args.chrome_profile_directory} ({args.chrome_user_data_dir})"
        )

    # 账号池不需要在这里通过 _ensure_login 打开浏览器，因为多个账号需要用户手动导入 cookie。
    asyncio.run(_ensure_login())

    uvicorn.run(app, host=args.host, port=args.port, log_level="info")
