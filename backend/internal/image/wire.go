package image

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewPromptCatalogRepository,
	NewPromptCatalogService,
	NewTwitterImportService,
)
