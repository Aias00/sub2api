const IMAGE_GENERATOR_DRAFT_KEY = 'sub2api:image-generator:draft'

export interface ImageGeneratorDraft {
  prompt: string
  title?: string
  sourcePromptId?: string
  source?: string
}

export function saveImageGeneratorDraft(draft: ImageGeneratorDraft) {
  window.sessionStorage.setItem(IMAGE_GENERATOR_DRAFT_KEY, JSON.stringify(draft))
}

export function loadImageGeneratorDraft(): ImageGeneratorDraft | null {
  const rawDraft = window.sessionStorage.getItem(IMAGE_GENERATOR_DRAFT_KEY)
  if (!rawDraft) return null

  const draft = JSON.parse(rawDraft) as Partial<ImageGeneratorDraft>
  return typeof draft.prompt === 'string' && draft.prompt.trim()
    ? {
        prompt: draft.prompt,
        title: typeof draft.title === 'string' ? draft.title : undefined,
        sourcePromptId: typeof draft.sourcePromptId === 'string' ? draft.sourcePromptId : undefined,
        source: typeof draft.source === 'string' ? draft.source : undefined,
      }
    : null
}

export function clearImageGeneratorDraft() {
  window.sessionStorage.removeItem(IMAGE_GENERATOR_DRAFT_KEY)
}
