# rclone FTP 字符编码支持使用说明

## 概述

本版本 rclone 新增了 FTP 后端的字符编码转换功能，支持连接使用非 UTF-8 编码的 FTP 服务器（如中文 GBK、日文 Shift_JIS、韩文 EUC-KR 等），使本地系统能够正确显示文件名中的非 ASCII 字符。

## 使用方法

### 方法一：使用命令行参数（临时使用）

```bash
# 列出 FTP 服务器上的文件（测试用）
./rclone-new ls --ftp-host=192.168.50.134 --ftp-user=smtrix --ftp-pass=12345678 --ftp-charset=GBK :ftp:

# 挂载 FTP 到本地目录
mkdir -p /mnt/ftp_share
./rclone-new mount \
  --ftp-host=192.168.50.134 \
  --ftp-user=smtrix \
  --ftp-pass=12345678 \
  --ftp-charset=GBK \
  --ftp-concurrency=5 \
  --vfs-cache-mode=full \
  --allow-non-empty \
  --allow-other \
  --daemon \
  :ftp: /mnt/ftp_share
```

### 方法二：使用 rclone config 配置（推荐长期使用）

#### 交互式配置

```bash
./rclone-new config
```

按照提示操作：
1. 选择 `n` 新建 remote
2. 名称输入：`myftp`
3. 类型选择：`ftp`
4. host 输入：`192.168.50.134`
5. user 输入：`smtrix`
6. pass 输入：`12345678`
7. port 默认 `21`
8. **charset 输入：`GBK`** ← 关键参数！
9. 其他选项保持默认
10. 确认并退出

配置完成后使用：

```bash
# 列出文件
./rclone-new ls myftp:

# 挂载
mkdir -p /mnt/ftp_share
./rclone-new mount myftp: /mnt/ftp_share \
  --vfs-cache-mode=full \
  --allow-non-empty \
  --allow-other \
  --daemon
```

#### 直接编辑配置文件

编辑 `~/.config/rclone/rclone.conf`，添加以下内容：

```ini
[myftp]
type = ftp
host = 192.168.50.134
user = smtrix
pass = <加密后的密码>
port = 21
charset = GBK
```

### 方法三：使用环境变量（临时测试）

```bash
export RCLONE_FTP_HOST=192.168.50.134
export RCLONE_FTP_USER=smtrix
export RCLONE_FTP_PASS=12345678
export RCLONE_FTP_CHARSET=GBK

./rclone-new ls :ftp:
```

## 实际场景示例

### 场景：挂载 GBK 编码的 FTP 服务器

FTP 服务器信息：
- 地址：192.168.50.134
- 用户名：smtrix
- 密码：12345678
- 服务器编码：GBK

```bash
# 1. 创建挂载点
sudo mkdir -p /mnt/ftp_gbk

# 2. 挂载（前台运行，方便查看日志）
./rclone-new mount \
  --ftp-host=192.168.50.134 \
  --ftp-user=smtrix \
  --ftp-pass=12345678 \
  --ftp-charset=GBK \
  --ftp-concurrency=5 \
  --vfs-cache-mode=full \
  --allow-non-empty \
  --allow-other \
  --dir-cache-time=10m \
  --log-level=INFO \
  :ftp: /mnt/ftp_gbk

# 3. 在另一个终端中访问
ls -la /mnt/ftp_gbk
# 现在应该能看到正确显示的中文文件名了！

# 4. 卸载
fusermount -u /mnt/ftp_gbk
```

### 场景：复制 GBK 编码 FTP 上的文件到本地

```bash
# 复制整个目录
./rclone-new copy \
  --ftp-host=192.168.50.134 \
  --ftp-user=smtrix \
  --ftp-pass=12345678 \
  --ftp-charset=GBK \
  --progress \
  :ftp:/远程目录 /本地目录

# 同步（双向）
./rclone-new sync \
  --ftp-host=192.168.50.134 \
  --ftp-user=smtrix \
  --ftp-pass=12345678 \
  --ftp-charset=GBK \
  --progress \
  :ftp:/远程目录 /本地目录
```

## 支持的字符编码

| 编码名称 | 适用语言 | 示例 |
|---------|---------|------|
| `GBK` | 简体中文 | 中文文件名 |
| `Big5` | 繁体中文 | 中文文件名 |
| `Shift_JIS` | 日文 | 日本語ファイル名 |
| `EUC-JP` | 日文 | 日本語ファイル名 |
| `EUC-KR` | 韩文 | 한글파일명 |
| `ISO-8859-1` | 西欧语言 | ñóúü |
| `KOI8-R` | 俄文 | Русский |
| `windows-1251` | 俄文 | Русский |
| `GB18030` | 中文 | 中文文件名 |

## 常见问题

### Q: 设置 charset 后文件名还是乱码？
A: 请确认 FTP 服务器实际使用的编码。可以尝试不同的编码，如 `GBK`、`GB18030`、`Big5` 等。

### Q: 如何查看 FTP 服务器的编码？
A: 连接 FTP 服务器后，查看文件列表中的中文文件名是否正常显示。如果显示乱码，尝试切换不同的 charset 参数。

### Q: 上传文件时文件名会转换吗？
A: 会的。设置 charset 后，上传时本地 UTF-8 文件名会自动转换为服务器编码，下载时服务器编码的文件名会自动转换为本地 UTF-8。

### Q: 为什么需要设置 `--ftp-concurrency`？
A: FTP 协议在传输文件时会占用控制连接，设置并发数可以提高传输效率。建议设置为 5 左右。

## 调试技巧

如果遇到问题，可以开启调试日志：

```bash
./rclone-new --log-level=DEBUG \
  --ftp-host=192.168.50.134 \
  --ftp-user=smtrix \
  --ftp-pass=12345678 \
  --ftp-charset=GBK \
  ls :ftp:
```

查看日志中的 `FTP Tx` 和 `FTP Rx` 输出，确认字符编码是否正确转换。