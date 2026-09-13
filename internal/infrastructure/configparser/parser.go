package configparser

import (
	"bufio"
	"io"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/config"
)

// Parse 从 reader 中逐行读取并构建 ConfigDocument。
// 行号从 1 开始；无法解析的行被容错保留为 Comment 类型。
// 支持 UTF-8 BOM 跳过。
func Parse(reader io.Reader) (*config.ConfigDocument, error) {
	doc := &config.ConfigDocument{}
	br := bufio.NewReader(reader)
	lineNumber := 0
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			lineNumber++
			// 跳过 BOM（仅第一行）
			if lineNumber == 1 {
				line = strings.TrimPrefix(line, "\uFEFF")
			}
			// 检测末尾换行：以最近读取的一行是否以 \n 结束为准
			doc.EndsWithNewline = strings.HasSuffix(line, "\n")
			// 去除行尾 \n，但保留 \r 于 RawText 中以实现 CRLF 无损往返；
			// lex 时再去除 \r 以免污染 value。
			content := strings.TrimSuffix(line, "\n")
			node := lexLine(strings.TrimSuffix(content, "\r"), lineNumber)
			node.RawText = content
			doc.Nodes = append(doc.Nodes, node)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return doc, nil
}
