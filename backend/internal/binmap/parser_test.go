package binmap

import (
	"strings"
	"testing"
)

func TestParseFullFormat(t *testing.T) {
	content := `
# 示例 wafer map
Notch: down
BinDef: 1,Pass,#2ecc71,pass
BinDef: 2,Fail,#e74c3c,fail
DieSize: 5000,5000
Grid:
. . 1 1 .
. 1 1 2 .
. . 1 2 .
. . . . .
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if p.Rows != 4 {
		t.Fatalf("期望 4 行数据（含全空轮廓行），得到 %d", p.Rows)
	}
	if p.Cols != 5 {
		t.Fatalf("期望 5 列，得到 %d", p.Cols)
	}
	if p.Notch != "down" {
		t.Fatalf("notch 期望 down，得到 %s", p.Notch)
	}
	// (0,2)=1, (1,3)=2
	if p.Data[0][2] != 1 || p.Data[1][3] != 2 {
		t.Fatalf("网格内容不正确: %v", p.Data)
	}
	if p.Data[0][0] != -1 {
		t.Fatalf("空位应为 -1，得到 %d", p.Data[0][0])
	}
	if len(p.BinDefs) != 2 {
		t.Fatalf("期望 2 个 bin 定义，得到 %d", len(p.BinDefs))
	}
	if !p.BinDefs[0].Pass || p.BinDefs[1].Pass {
		t.Fatal("bin pass 标记不正确")
	}
}

func TestParseCSVGridAutoBins(t *testing.T) {
	content := "1,1,3\n1,3,-\n- - -\n"
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if p.Rows != 3 { // 全空行保留以维持晶圆轮廓
		t.Fatalf("期望 3 行，得到 %d, %v", p.Rows, p.Data)
	}
	if p.Cols != 3 {
		t.Fatalf("期望 3 列，得到 %d", p.Cols)
	}
	if p.Data[1][2] != -1 {
		t.Fatalf("空位应为 -1，得到 %d", p.Data[1][2])
	}
	bins := map[int]bool{}
	for _, b := range []int{1, 3} {
		bins[b] = true
	}
	for _, d := range p.BinDefs {
		if !bins[d.Number] {
			t.Fatalf("出现未预期 bin: %d", d.Number)
		}
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := Parse(""); err == nil {
		t.Fatal("空文件应报错")
	}
	if _, err := Parse("Notch: north\n1 1\n"); err == nil {
		t.Fatal("非法 notch 应报错")
	}
	if _, err := Parse(strings.Repeat("1 ", 1100)); err == nil {
		t.Fatal("超过最大网格应报错")
	}
}
