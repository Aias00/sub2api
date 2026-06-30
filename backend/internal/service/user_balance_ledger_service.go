package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// EntryType 流水类型枚举
type BalanceLedgerEntryType string

const (
	EntryTypeRecharge         BalanceLedgerEntryType = "recharge"          // 充值订单
	EntryTypeAPIUsage         BalanceLedgerEntryType = "api_usage"         // API调用扣费
	EntryTypeImageWorkspace   BalanceLedgerEntryType = "image_workspace"   // 图片工作台扣费
	EntryTypeWechatExport     BalanceLedgerEntryType = "wechat_export"     // 微信导出扣费
	EntryTypeRedeem           BalanceLedgerEntryType = "redeem"            // 兑换码兑换
	EntryTypeAdminAdjustment  BalanceLedgerEntryType = "admin_adjustment"  // 管理员调整
	EntryTypeAffiliateTransfer BalanceLedgerEntryType = "affiliate_transfer" // 返利转余额
	EntryTypeRefund           BalanceLedgerEntryType = "refund"            // 退款扣减
	EntryTypePromoBonus       BalanceLedgerEntryType = "promo_bonus"       // 优惠码奖励
	EntryTypeOAuthBindBonus   BalanceLedgerEntryType = "oauth_bind_bonus"  // OAuth首次绑定奖励
	EntryTypeExpiry           BalanceLedgerEntryType = "expiry"            // 过期清零
	EntryTypeCorrection       BalanceLedgerEntryType = "correction"        // 系统纠正
)

// SourceType 来源类型枚举
type BalanceLedgerSourceType string

const (
	SourceTypePaymentOrder       BalanceLedgerSourceType = "payment_order"
	SourceTypeUsageLog           BalanceLedgerSourceType = "usage_log"
	SourceTypeRedeemCode         BalanceLedgerSourceType = "redeem_code"
	SourceTypeAdminAction        BalanceLedgerSourceType = "admin_action"
	SourceTypeAffiliateLedger    BalanceLedgerSourceType = "affiliate_ledger"
	SourceTypeRefund             BalanceLedgerSourceType = "refund"
	SourceTypePromoCodeUsage     BalanceLedgerSourceType = "promo_code_usage"
	SourceTypeOAuthBinding       BalanceLedgerSourceType = "oauth_binding"
	SourceTypeImageWorkspaceRecord BalanceLedgerSourceType = "image_workspace_record"
	SourceTypeWechatExportTask   BalanceLedgerSourceType = "wechat_export_task"
	SourceTypeSystemCorrection   BalanceLedgerSourceType = "system_correction"
)

// UserBalanceLedgerEntry 余额流水记录
type UserBalanceLedgerEntry struct {
	ID            int64              `json:"id"`
	UserID        int64              `json:"user_id"`
	EntryType     BalanceLedgerEntryType `json:"entry_type"`
	Amount        float64            `json:"amount"`          // 正数入账，负数扣费
	BalanceBefore *float64           `json:"balance_before"`  // 变动前余额（历史数据可能为 NULL）
	BalanceAfter  *float64           `json:"balance_after"`   // 变动后余额
	SourceType    BalanceLedgerSourceType `json:"source_type"`
	SourceID      *int64             `json:"source_id"`       // 来源记录 ID
	Description   string             `json:"description"`
	MetadataJSON  json.RawMessage    `json:"metadata_json"`   // 扩展元数据
	CreatedAt     time.Time          `json:"created_at"`
}

// BalanceLedgerFilter 流水查询过滤器
type BalanceLedgerFilter struct {
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	EntryTypes []BalanceLedgerEntryType `json:"entry_types"` // 筛选类型（可选）
	StartAt    *time.Time             `json:"start_at"`    // 开始时间
	EndAt      *time.Time             `json:"end_at"`      // 结束时间
}

// UserBalanceLedgerRepository 余额流水仓储接口
type UserBalanceLedgerRepository interface {
	// Create 写入流水记录（在事务内调用）
	Create(ctx context.Context, entry *UserBalanceLedgerEntry) error

	// CreateTx 在指定事务内写入流水记录（ExecContext 返回 sql.Result）
	CreateTx(ctx context.Context, exec interface{ ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) }, entry *UserBalanceLedgerEntry) error

	// ListByUser 查询用户流水（分页）
	ListByUser(ctx context.Context, userID int64, filter BalanceLedgerFilter) ([]UserBalanceLedgerEntry, int64, error)

	// GetBySource 查询某来源的流水（用于去重检查）
	GetBySource(ctx context.Context, sourceType BalanceLedgerSourceType, sourceID int64) (*UserBalanceLedgerEntry, error)
}

// UserBalanceLedgerService 余额流水服务
type UserBalanceLedgerService struct {
	ledgerRepo UserBalanceLedgerRepository
	userRepo   UserRepository
}

// NewUserBalanceLedgerService 创建余额流水服务
func NewUserBalanceLedgerService(
	ledgerRepo UserBalanceLedgerRepository,
	userRepo UserRepository,
) *UserBalanceLedgerService {
	return &UserBalanceLedgerService{
		ledgerRepo: ledgerRepo,
		userRepo:   userRepo,
	}
}

// WriteLedger 写入余额流水
// 参数说明：
//   - ctx: 上下文（可能包含事务）
//   - userID: 用户 ID
//   - entryType: 流水类型
//   - amount: 金额变动（正数入账，负数扣费）
//   - balanceBefore: 变动前余额（可选，若为 nil 则不记录）
//   - sourceType: 来源类型
//   - sourceID: 来源记录 ID（可选）
//   - description: 流水描述
//   - metadata: 扩展元数据（可选）
func (s *UserBalanceLedgerService) WriteLedger(
	ctx context.Context,
	userID int64,
	entryType BalanceLedgerEntryType,
	amount float64,
	balanceBefore *float64,
	sourceType BalanceLedgerSourceType,
	sourceID *int64,
	description string,
	metadata map[string]interface{},
) error {
	// 构建 metadata JSON
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	entry := &UserBalanceLedgerEntry{
		UserID:        userID,
		EntryType:     entryType,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		SourceType:    sourceType,
		SourceID:      sourceID,
		Description:   description,
		MetadataJSON:  metadataJSON,
		CreatedAt:     time.Now(),
	}

	return s.ledgerRepo.Create(ctx, entry)
}

// WriteLedgerTx 在事务内写入余额流水，并记录余额前后值
// 此方法在余额变动的事务内调用，确保流水与余额变动原子性
func (s *UserBalanceLedgerService) WriteLedgerTx(
	ctx context.Context,
	exec interface{ ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) },
	userID int64,
	entryType BalanceLedgerEntryType,
	amount float64,
	balanceBefore float64,
	balanceAfter float64,
	sourceType BalanceLedgerSourceType,
	sourceID *int64,
	description string,
	metadata map[string]interface{},
) error {
	// 构建 metadata JSON
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	entry := &UserBalanceLedgerEntry{
		UserID:        userID,
		EntryType:     entryType,
		Amount:        amount,
		BalanceBefore: &balanceBefore,
		BalanceAfter:  &balanceAfter,
		SourceType:    sourceType,
		SourceID:      sourceID,
		Description:   description,
		MetadataJSON:  metadataJSON,
		CreatedAt:     time.Now(),
	}

	return s.ledgerRepo.CreateTx(ctx, exec, entry)
}

// ListUserLedger 查询用户余额流水
func (s *UserBalanceLedgerService) ListUserLedger(
	ctx context.Context,
	userID int64,
	filter BalanceLedgerFilter,
) ([]UserBalanceLedgerEntry, int64, error) {
	// 设置默认分页
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	return s.ledgerRepo.ListByUser(ctx, userID, filter)
}

// EntryTypeDisplayName 流水类型的显示名称
func EntryTypeDisplayName(entryType BalanceLedgerEntryType) string {
	names := map[BalanceLedgerEntryType]string{
		EntryTypeRecharge:          "充值",
		EntryTypeAPIUsage:          "API调用扣费",
		EntryTypeImageWorkspace:    "图片生成扣费",
		EntryTypeWechatExport:      "微信导出扣费",
		EntryTypeRedeem:            "兑换码兑换",
		EntryTypeAdminAdjustment:   "管理员调整",
		EntryTypeAffiliateTransfer: "返利转入",
		EntryTypeRefund:            "退款扣减",
		EntryTypePromoBonus:        "优惠码奖励",
		EntryTypeOAuthBindBonus:    "绑定奖励",
		EntryTypeExpiry:            "过期清零",
		EntryTypeCorrection:        "系统纠正",
	}
	if name, ok := names[entryType]; ok {
		return name
	}
	return string(entryType)
}

// EntryTypeIcon 流水类型对应的图标名称
func EntryTypeIcon(entryType BalanceLedgerEntryType) string {
	icons := map[BalanceLedgerEntryType]string{
		EntryTypeRecharge:          "dollar-up",
		EntryTypeAPIUsage:          "zap",
		EntryTypeImageWorkspace:    "image",
		EntryTypeWechatExport:      "wechat",
		EntryTypeRedeem:            "gift",
		EntryTypeAdminAdjustment:   "settings",
		EntryTypeAffiliateTransfer: "users",
		EntryTypeRefund:            "dollar-down",
		EntryTypePromoBonus:        "tag",
		EntryTypeOAuthBindBonus:    "link",
		EntryTypeExpiry:            "clock",
		EntryTypeCorrection:        "wrench",
	}
	if icon, ok := icons[entryType]; ok {
		return icon
	}
	return "circle"
}

// EntryTypeColor 流水类型对应的颜色（CSS class）
func EntryTypeColor(entryType BalanceLedgerEntryType) string {
	colors := map[BalanceLedgerEntryType]string{
		EntryTypeRecharge:          "green",    // 入账
		EntryTypeAPIUsage:          "red",      // 扣费
		EntryTypeImageWorkspace:    "red",      // 扣费
		EntryTypeWechatExport:      "red",      // 扣费
		EntryTypeRedeem:            "blue",     // 兑换
		EntryTypeAdminAdjustment:   "yellow",   // 调整（可能是正或负）
		EntryTypeAffiliateTransfer: "purple",   // 转入
		EntryTypeRefund:            "red",      // 扣减
		EntryTypePromoBonus:        "green",    // 奖励
		EntryTypeOAuthBindBonus:    "green",    // 奖励
		EntryTypeExpiry:            "gray",     // 过期
		EntryTypeCorrection:        "orange",   // 纠正
	}
	if color, ok := colors[entryType]; ok {
		return color
	}
	return "gray"
}