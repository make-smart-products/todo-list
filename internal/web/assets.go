package web

import "embed"

// Assets contains the browser client for Oil Worker.
//
//go:embed app/*
var Assets embed.FS
