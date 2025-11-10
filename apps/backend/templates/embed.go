package templates

import "embed"

//go:embed emails/*
var TemplateDir embed.FS
