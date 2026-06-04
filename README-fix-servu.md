# rclone FTP 后端 Serv-U v10.3 兼容性修复

## 问题描述

Serv-U FTP v10.3（Windows + GBK 编码）服务器无法识别带双引号的 LIST 命令路径。

### 症状
- rclone 自动发送：`LIST "中文 目录名"`（带双引号）
- Serv-U v10.3 不识别带引号路径，返回 0 字节
- 目录无法进入，中文带空格目录访问失败
- curlftpfs 正常（发送 `LIST 中文 目录名` 不带双引号）

### 根源
`jlaffaye/ftp` 库中的 `quotePathInCommand()` 函数在路径含空格/中文时自动加双引号。

## 解决方案

新增 FTP 后端配置项：**`fix_servu_list_quotes`**

- `fix_servu_list_quotes = true` → 开启 Serv-U 兼容模式（不加引号）
- `fix_servu_list_quotes = false`（默认）→ 保持 rclone 原始行为

## 修改的文件

| 文件 | 修改内容 |
|------|---------|
| `vendor/github.com/jlaffaye/ftp/conn.go` | `Options` 结构体新增 `DisablePathQuoting` 字段；`quotePathInCommand()` 函数新增 `disableQuoting` 参数 |
| `vendor/github.com/jlaffaye/ftp/ftp.go` | `dialOptions` 结构体新增 `disablePathQuoting` 字段；`cmd()` 和 `cmdDataConnFrom()` 方法传递开关值 |
| `backend/ftp/ftp.go` | `Options` 结构体新增 `FixServuListQuotes` 字段；`init()` 中注册配置项；`ftpConnection()` 中传递配置 |

## 编译方法（Ubuntu 22.04）

```bash
cd /path/to/rclone
go build -o rclone ./
```

或使用 make：

```bash
cd /path/to/rclone
make
```

## 配置示例

在 `rclone.conf` 中添加：

```ini
[servu]
type = ftp
host = 192.168.1.100
user = username
pass = password
charset = GBK
fix_servu_list_quotes = true
```

## 测试方法

```bash
# 列出根目录
./rclone lsd servu:

# 进入中文带空格目录
./rclone lsd "servu:/中文 目录名"

# 同步测试
./rclone sync "servu:/中文 目录名" /local/test
```

## 设计原则

1. **最小侵入式**：只修改 3 个文件，新增约 20 行代码
2. **默认行为不变**：`fix_servu_list_quotes = false` 时完全保持 rclone 原始行为
3. **不影响其他 FTP 服务器**：只有显式开启此选项才会改变行为
4. **保留所有原有功能**：GBK、charset、encoding、disable_utf8 等全部不受影响
5. **开关可随时切换**：无需重新编译，修改配置文件即可

## 兼容性

| 场景 | fix_servu_list_quotes=false | fix_servu_list_quotes=true |
|------|:---:|:---:|
| 普通 FTP 服务器（FileZilla、ProFTPD 等） | ✅ 正常 | ✅ 正常 |
| Serv-U v10.3（GBK + 中文空格目录） | ❌ 失败 | ✅ 正常 |
| 其他 FTP 服务器 | ✅ 正常 | ✅ 正常 |
