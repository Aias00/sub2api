# Business Capabilities Guide

cloudbase is not limited to API keys, model groups, and billing. It also provides product-facing workflows for content operations and image generation. The current business capabilities include WeChat export, hot topic tracking, image prompt catalog, image generation workspace, and My Tasks.

Business entries: [Home](/home) / [WeChat Export](/wechat) / [Hot Topics](/hot) / [Image Prompts](/prompts) / [Image Workspace](/image-generator) / [My Tasks](/tasks)

## Capability Overview

### WeChat Export

WeChat Export turns public-account articles into downloadable artifacts. It is useful for article archiving, material collection, operation handoff, and moving public-account content into later editing workflows.

Core capabilities:

- Bind or sync public-account articles.
- Select articles and create export tasks.
- Export as HTML, Markdown, JSON, or configured format combinations.
- Download artifacts from the task list after completion.

### Hot Topics

Hot Topics provides a focused stream of recent topics. It is intended for topic discovery, operation monitoring, and content planning before production.

The page focuses on the main hot-topic feed. Browse titles, sources, and timestamps, then decide whether a topic should enter the content or image production flow.

### Image Prompts

The Image Prompt Catalog stores reusable image-generation prompts. It helps turn examples, style templates, and structured prompt patterns into reusable team assets.

Common uses:

- Search existing prompts.
- Filter by tags, source, model, or topic.
- Copy prompts for external tools.
- Jump directly into the image generation workspace.

### Image Workspace

The Image Workspace creates real image generation tasks and manages generated artifacts. It is backed by task queues instead of being a one-off frontend demo.

Core capabilities:

- Enter prompt, negative prompt, and style notes.
- Choose model, size, quality, and batch count.
- Create an image task and wait for worker processing.
- Review task status, preview images, and download originals.
- Retry failed tasks or cancel queued tasks.

### My Tasks

My Tasks is the unified task center for WeChat export and image generation jobs.

Use it to:

- View all, active, completed, and attention-needed tasks.
- Filter by WeChat export or image generation.
- Open the original business page for context.
- Download completed artifacts.
- Retry failed or retryable tasks.

## Recommended Workflows

### Archive Content: WeChat Articles to Local Materials

1. Open [WeChat Export](/wechat).
2. Sign in and sync public-account articles.
3. Search for or select the articles you need.
4. Choose export formats and create a task.
5. Open [My Tasks](/tasks) and wait for completion.
6. Download the export package for editing, archiving, or publishing.

If a direct article import hits a WeChat verification page, use public-account sync instead. The direct import flow will report the verification-page problem instead of storing that page as an article.

### Topic Production: Hot Topic to Prompt to Image

1. Open [Hot Topics](/hot) and browse recent topics.
2. Note suitable themes, keywords, and angles.
3. Open [Image Prompts](/prompts) and search for related styles or topics.
4. Copy a prompt or choose the image-generation action to open [Image Workspace](/image-generator).
5. Adjust the prompt for the selected topic and create an image task.
6. Track status and download finished images in [My Tasks](/tasks).

### Image Production: Template to Artifact

1. Open [Image Prompts](/prompts).
2. Pick a prompt close to your desired style.
3. Use the image-generation action to send it into [Image Workspace](/image-generator).
4. Adjust model, size, quality, and batch count.
5. Submit the task and wait for generation.
6. Preview the result and download the original image.

## Task Statuses

The task center normalizes statuses across business workflows:

- **Queued**: the task has been created and is waiting for a worker.
- **Running**: a worker is processing the task.
- **Completed**: the task succeeded and artifacts can be opened or downloaded.
- **Partial**: some articles or artifacts succeeded while others failed.
- **Failed**: the task did not complete; inspect the error and retry or adjust input.
- **Cancelled**: the task was cancelled and will not continue.

## Usage Tips

- Keep task context clear, such as public account, article topic, or date.
- For image prompts, specify subject, style, composition, size, and constraints.
- When a task fails, read the task error before retrying.
- Download artifacts only after the task is completed or partially completed.
- Never expose account credentials, cookies, API keys, or public-account sensitive data in screenshots or support messages.

## Related Pages

- [WeChat Export](/wechat)
- [Hot Topics](/hot)
- [Image Prompts](/prompts)
- [Image Workspace](/image-generator)
- [My Tasks](/tasks)
