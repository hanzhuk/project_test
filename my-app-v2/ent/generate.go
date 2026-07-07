// Package ent 提供 Ent ORM 的代码生成入口。
// 执行 `go generate ./ent` 或 `make ent-gen` 可根据 schema 重新生成代码。
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
