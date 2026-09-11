package binmap

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"sample-tracker/internal/model"
)

const (
	maxGridDim = 1000
)

// ParsedMap 解析结果
type ParsedMap struct {
	Data    [][]int
	Rows    int
	Cols    int
	Notch   string
	BinDefs []model.BinDef
}

// 默认调色板（未在 BinDef 中声明的 bin 依次取色）
var palette = []string{
	"#2ecc71", "#e74c3c", "#f39c12", "#3498db", "#9b59b6",
	"#1abc9c", "#e67e22", "#34495e", "#e84393", "#00cec9",
	"#6c5ce7", "#fd79a8", "#00b894", "#d63031", "#0984e3",
}

// Parse 解析文本 bin map。
//
// 支持头部（顺序无关，均可省略）：
//
//	Notch: down|up|left|right
//	BinDef: <bin号>,<名称>,<#颜色>,<pass|fail>
//
// 网格支持两种写法：出现 "Grid:" 标记后开始；或直接是数据行。
// 行内用空白或逗号分隔；. - x X na none null 表示无晶粒，整数为 bin 号。
func Parse(content string) (*ParsedMap, error) {
	sc := bufio.NewScanner(strings.NewReader(content))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	p := &ParsedMap{Notch: "down"}
	var rows [][]int
	gridStarted := false
	seenBin := map[int]bool{}
	binOrder := []int{}
	defByBin := map[int]model.BinDef{}

	for sc.Scan() {
		raw := strings.TrimSpace(sc.Text())
		if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "//") {
			continue
		}

		lower := strings.ToLower(raw)
		if !gridStarted {
			switch {
			case strings.HasPrefix(lower, "notch:"):
				n := strings.ToLower(strings.TrimSpace(raw[len("notch:"):]))
				switch n {
				case "up", "down", "left", "right":
					p.Notch = n
				default:
					return nil, fmt.Errorf("未知的 Notch 方向: %s", n)
				}
				continue
			case strings.HasPrefix(lower, "bindef:"):
				d, err := parseBinDef(raw[len("bindef:"):])
				if err != nil {
					return nil, err
				}
				defByBin[d.Number] = d
				continue
			case strings.HasPrefix(lower, "diesize:"):
				continue // 晶粒尺寸仅作记录，可视化自适应，忽略
			case strings.HasPrefix(lower, "grid:"):
				gridStarted = true
				rest := strings.TrimSpace(raw[len("grid:"):])
				if rest == "" {
					continue
				}
				raw = rest
			}
		} else if strings.HasSuffix(lower, ":grid") || lower == "end" || lower == "endgrid" {
			break
		}

		row, ok := parseGridRow(raw, seenBin, &binOrder)
		if !ok {
			if len(rows) == 0 && !gridStarted {
				// 头部出现无法识别的行，忽略
				continue
			}
			return nil, fmt.Errorf("无法解析的网格行: %q", raw)
		}
		rows = append(rows, row)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("文件中没有网格数据")
	}

	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}
	if cols == 0 || len(rows) > maxGridDim || cols > maxGridDim {
		return nil, fmt.Errorf("网格尺寸非法（最大 %d×%d）", maxGridDim, maxGridDim)
	}
	// 短行补 -1
	for i := range rows {
		for len(rows[i]) < cols {
			rows[i] = append(rows[i], -1)
		}
	}

	// 合并 bin 定义：显式声明 + 文件中出现但未声明的 bin（自动配色）
	defs := make([]model.BinDef, 0, len(binOrder))
	autoIdx := 0
	for _, b := range binOrder {
		if d, ok := defByBin[b]; ok {
			defs = append(defs, d)
			continue
		}
		color := palette[autoIdx%len(palette)]
		autoIdx++
		defs = append(defs, model.BinDef{
			Number: b,
			Name:   fmt.Sprintf("Bin%d", b),
			Color:  color,
			Pass:   b == 1, // 约定：未声明时 1 为合格 bin
		})
	}
	// 文件中声明了但网格里没有的 bin 也保留
	for _, d := range defByBin {
		if !seenBin[d.Number] {
			defs = append(defs, d)
		}
	}

	p.Data = rows
	p.Rows = len(rows)
	p.Cols = cols
	p.BinDefs = defs
	return p, nil
}

func parseBinDef(s string) (model.BinDef, error) {
	parts := strings.Split(s, ",")
	if len(parts) < 1 {
		return model.BinDef{}, fmt.Errorf("BinDef 至少需要 bin 号")
	}
	n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || n < 0 {
		return model.BinDef{}, fmt.Errorf("BinDef bin 号非法: %q", strings.TrimSpace(parts[0]))
	}
	d := model.BinDef{Number: n, Name: fmt.Sprintf("Bin%d", n), Color: palette[n%len(palette)], Pass: n == 1}
	if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
		d.Name = strings.TrimSpace(parts[1])
	}
	if len(parts) >= 3 && strings.TrimSpace(parts[2]) != "" {
		c := strings.TrimSpace(parts[2])
		if strings.HasPrefix(c, "#") || strings.HasPrefix(c, "rgb") {
			d.Color = c
		}
	}
	if len(parts) >= 4 {
		switch strings.ToLower(strings.TrimSpace(parts[3])) {
		case "pass", "good", "ok", "1":
			d.Pass = true
		case "fail", "ng", "bad", "0":
			d.Pass = false
		}
	}
	return d, nil
}

var emptyTokens = map[string]bool{
	".": true, "-": true, "x": true, "xx": true, "na": true,
	"n/a": true, "none": true, "null": true, "": true,
}

func parseGridRow(raw string, seen map[int]bool, order *[]int) ([]int, bool) {
	line := strings.ReplaceAll(raw, ",", " ")
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, true
	}
	row := make([]int, 0, len(fields))
	dieCount := 0
	for _, f := range fields {
		if emptyTokens[strings.ToLower(f)] {
			row = append(row, -1)
			continue
		}
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, false
		}
		if n < 0 || n > 9999 {
			return nil, false
		}
		row = append(row, n)
		dieCount++
		if !seen[n] {
			seen[n] = true
			*order = append(*order, n)
		}
	}
	if dieCount == 0 {
		return nil, true // 整行都是空位，跳过
	}
	return row, true
}
