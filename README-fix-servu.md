# Rclone Serv-U v10 FTP 修复说明

## 版本信息
- **版本号**: v1.75.0-dev
- **基于**: rclone v1.75.0
- **修复日期**: 2025-06-04

## 修复内容

### 问题描述
在访问 Serv-U FTP Server v10.3 时，当目录名包含中文和空格（如 "13 公用上传文件夹"），rclone 无法进入子目录。原因是 Serv-U v10 不支持 LIST 命令中带双引号的路径参数。

### 问题根因
`vendor/github.com/jlaffaye/ftp` 库中，`cmd()` 和 `cmdDataConnFrom()` 方法只检查了 `c.options.disablePathQuoting`（连接建立时设置），但没有检查 `c.ftpOptions.DisablePathQuoting`（通过 `SetOptions` 登录后设置）。导致 `fix_servu_list_quotes=true` 配置虽然传到底层库，但实际生成 LIST 命令时未生效。

### 修改文件
- `vendor/github.com/jlaffaye/ftp/ftp.go` — 底层 FTP 客户端库

### 修改内容
在 `cmd()` 和 `cmdDataConnFrom()` 方法中，增加对 `c.ftpOptions.DisablePathQuoting` 的检查：

```go
// 修改前：
line = quotePathInCommand(line, c.options.disablePathQuoting)

// 修改后：
disableQuoting := c.options.disablePathQuoting
if c.ftpOptions != nil && c.ftpOptions.DisablePathQuoting {
    disableQuoting = true
}
line = quotePathInCommand(line, disableQuoting)
```

### 效果
- `--ftp-fix-servu-list-quotes=true` 时：**所有** FTP 命令（LIST、CWD、RETR、STOR 等）的路径参数都不加双引号
- `--ftp-fix-servu-list-quotes=false`（默认）：完全保留原生逻辑，不影响其他用户

### 使用方式
```bash
# 命令行使用
./rclone lsd ftp:/ --ftp-fix-servu-list-quotes

# 或者在配置中设置
[remote]
type = ftp
host = your-servu-server
fix_servu_list_quotes = true
```

## 编译说明

### ARM64 架构
```bash
go build -buildvcs=false -o rclone-arm64
```

### x86_64 架构（交叉编译）
```bash
GOOS=linux GOARCH=amd64 go build -buildvcs=false -o rclonex84
```

## 相关文件
- `backend/ftp/ftp.go` — rclone FTP 后端（配置注册）
- `vendor/github.com/jlaffaye/ftp/ftp.go` — 底层 FTP 客户端库（修复位置）
- `vendor/github.com/jlaffaye/ftp/conn.go` — FTP 连接选项定义
