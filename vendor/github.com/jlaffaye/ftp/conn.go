package ftp

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// Options represents options that can be set on a ServerConn after login
type Options struct {
	// Charset is the character encoding to use for FTP commands and responses.
	// If empty, UTF-8 is used.
	Charset string
}

// DialWithTLSNoHandshake returns a DialOption that configures the ServerConn with specified TLS config
// but does not perform the TLS handshake immediately. The handshake will be performed on the first
// Read or Write. This is useful for FTP servers that hang during the TLS handshake.
func DialWithTLSNoHandshake(tlsConfig *tls.Config) DialOption {
	return DialOption{func(do *dialOptions) {
		do.tlsConfig = tlsConfig
		do.dialFunc = func(network, address string) (net.Conn, error) {
			conn, err := net.DialTimeout(network, address, do.dialer.Timeout)
			if err != nil {
				return nil, err
			}
			tlsConn := tls.Client(conn, tlsConfig)
			return tlsConn, nil
		}
	}}
}

// SetDeadline sets the read/write deadline on the underlying network connection.
func (c *ServerConn) SetDeadline(t time.Time) error {
	return c.netConn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline on the underlying network connection.
func (c *ServerConn) SetReadDeadline(t time.Time) error {
	return c.netConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline on the underlying network connection.
func (c *ServerConn) SetWriteDeadline(t time.Time) error {
	return c.netConn.SetWriteDeadline(t)
}

// SetOptions sets options on the ServerConn after login.
// Currently only Charset is supported.
func (c *ServerConn) SetOptions(opts *Options) {
	c.ftpOptions = opts
}

// quoteFTPPath quotes a path for use in FTP commands if it contains
// spaces or double quotes, per RFC 959.
//
// Rules:
//   - If the path contains no spaces and no double quotes, return as-is.
//   - Otherwise, double any embedded double quotes, then wrap the
//     entire path in double quotes.
func quoteFTPPath(path string) string {
	if path == "" {
		return path
	}
	// Check if quoting is needed
	needsQuoting := false
	for _, r := range path {
		if r == ' ' || r == '"' {
			needsQuoting = true
			break
		}
	}
	if !needsQuoting {
		return path
	}
	// Escape embedded double quotes by doubling them
	escaped := strings.ReplaceAll(path, `"`, `""`)
	// Wrap in double quotes
	return `"` + escaped + `"`
}

// getEncoding returns the Go encoding.Encoding for the given charset name.
// Returns nil if the charset is empty, "UTF-8", or unknown.
func getEncoding(charset string) encoding.Encoding {
	switch strings.ToUpper(charset) {
	case "", "UTF-8", "UTF8":
		return nil
	case "GBK", "CP936", "MS936":
		return simplifiedchinese.GBK
	case "GB18030":
		return simplifiedchinese.GB18030
	case "BIG5", "BIG-5", "CP950":
		return traditionalchinese.Big5
	case "SHIFT_JIS", "SHIFT-JIS", "SJIS", "CP932":
		return japanese.ShiftJIS
	case "EUC-JP", "EUCJP":
		return japanese.EUCJP
	case "EUC-KR", "EUCKR", "CP949":
		return korean.EUCKR
	case "ISO-8859-1", "LATIN1":
		return charmap.ISO8859_1
	case "ISO-8859-2", "LATIN2":
		return charmap.ISO8859_2
	case "ISO-8859-5":
		return charmap.ISO8859_5
	case "ISO-8859-6":
		return charmap.ISO8859_6
	case "ISO-8859-7":
		return charmap.ISO8859_7
	case "ISO-8859-8":
		return charmap.ISO8859_8
	case "ISO-8859-9":
		return charmap.ISO8859_9
	case "ISO-8859-10":
		return charmap.ISO8859_10
	case "ISO-8859-13":
		return charmap.ISO8859_13
	case "ISO-8859-14":
		return charmap.ISO8859_14
	case "ISO-8859-15":
		return charmap.ISO8859_15
	case "ISO-8859-16":
		return charmap.ISO8859_16
	case "KOI8-R":
		return charmap.KOI8R
	case "KOI8-U":
		return charmap.KOI8U
	case "WINDOWS-1250", "CP1250":
		return charmap.Windows1250
	case "WINDOWS-1251", "CP1251":
		return charmap.Windows1251
	case "WINDOWS-1252", "CP1252":
		return charmap.Windows1252
	case "WINDOWS-1253", "CP1253":
		return charmap.Windows1253
	case "WINDOWS-1254", "CP1254":
		return charmap.Windows1254
	case "WINDOWS-1255", "CP1255":
		return charmap.Windows1255
	case "WINDOWS-1256", "CP1256":
		return charmap.Windows1256
	case "WINDOWS-1257", "CP1257":
		return charmap.Windows1257
	case "WINDOWS-1258", "CP1258":
		return charmap.Windows1258
	case "IBM437", "CP437":
		return charmap.CodePage437
	case "IBM850", "CP850":
		return charmap.CodePage850
	case "IBM852", "CP852":
		return charmap.CodePage852
	case "IBM866", "CP866":
		return charmap.CodePage866
	case "MACINTOSH", "MAC":
		return charmap.Macintosh
	default:
		return nil
	}
}

// encodeUTF8ToCharset converts a UTF-8 string to the specified charset.
// If charset is empty or UTF-8, returns the original string unchanged.
func encodeUTF8ToCharset(s string, charset string) (string, error) {
	enc := getEncoding(charset)
	if enc == nil {
		return s, nil
	}
	// Transform from UTF-8 to target encoding
	result, _, err := transform.String(enc.NewEncoder(), s)
	if err != nil {
		return s, fmt.Errorf("failed to encode to %s: %w", charset, err)
	}
	return result, nil
}

// decodeCharsetToUTF8 converts a byte slice from the specified charset to UTF-8.
// If charset is empty or UTF-8, returns the original bytes unchanged.
func decodeCharsetToUTF8(data []byte, charset string) (string, error) {
	enc := getEncoding(charset)
	if enc == nil {
		return string(data), nil
	}
	// Transform from target encoding to UTF-8
	result, _, err := transform.String(enc.NewDecoder(), string(data))
	if err != nil {
		return string(data), fmt.Errorf("failed to decode from %s: %w", charset, err)
	}
	return result, nil
}

// quotePathInCommand applies quoteFTPPath to the path argument in an FTP
// command line. It handles commands like:
//
//	CWD, RETR, STOR, APPE, DELE, MKD, RMD, RNFR, RNTO, SIZE, MDTM, MFMT,
//	LIST, NLST, MLSD, STAT, XCWD, XMKD, XRMD
//
// The format is: "COMMAND path" or "COMMAND arg1 arg2" (for RNFR/RNTO
// the second token is the path; for MFMT the third token is the path).
func quotePathInCommand(line string) string {
	// Split into command and arguments
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return line
	}

	cmd := strings.ToUpper(parts[0])
	rest := parts[1]

	// Commands that take a single path argument
	singlePathCmds := map[string]bool{
		"CWD":   true,
		"RETR":  true,
		"STOR":  true,
		"APPE":  true,
		"DELE":  true,
		"MKD":   true,
		"RMD":   true,
		"PWD":   true,
		"LIST":  true,
		"NLST":  true,
		"MLSD":  true,
		"STAT":  true,
		"XCWD":  true,
		"XMKD":  true,
		"XRMD":  true,
		"SIZE":  true,
		"MDTM":  true,
		"CDUP":  true,
		"NOOP":  true,
		"QUIT":  true,
		"TYPE":  true,
		"PASS":  true,
		"USER":  true,
		"ACCT":  true,
		"SMNT":  true,
		"REIN":  true,
		"ALLO":  true,
		"SYST":  true,
		"HELP":  true,
		"ABOR":  true,
	}

	// Commands that take two path arguments (RNFR, RNTO)
	// For these, we need to quote each path argument separately
	if cmd == "RNFR" || cmd == "RNTO" {
		quoted := quoteFTPPath(rest)
		return cmd + " " + quoted
	}

	// MFMT takes: MFMT YYYYMMDDHHMMSS path
	if cmd == "MFMT" {
		// Split rest into time and path
		mfmtParts := strings.SplitN(rest, " ", 2)
		if len(mfmtParts) == 2 {
			timeStr := mfmtParts[0]
			pathStr := mfmtParts[1]
			quotedPath := quoteFTPPath(pathStr)
			return cmd + " " + timeStr + " " + quotedPath
		}
		return line
	}

	// For single path commands, quote the path argument
	if singlePathCmds[cmd] {
		quoted := quoteFTPPath(rest)
		return cmd + " " + quoted
	}

	// For unknown commands, try to quote the first argument if it looks like a path
	// (heuristic: contains a slash or non-ASCII characters)
	if strings.ContainsAny(rest, "/\\") || !isASCII(rest) {
		quoted := quoteFTPPath(rest)
		return cmd + " " + quoted
	}

	return line
}

// isASCII checks if a string contains only ASCII characters.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
