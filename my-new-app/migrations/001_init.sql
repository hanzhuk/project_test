-- my-new-app 数据库初始化脚本
-- PostgreSQL 15+

-- 创建数据库
CREATE DATABASE my-new-app_db;

-- 连接到数据库
\c my-new-app_db;

-- 创建用户表（Ent 自动迁移会自动创建，此脚本作为手动初始化参考）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
