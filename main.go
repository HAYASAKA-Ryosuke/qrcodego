package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
)

func drawCell(m *image.RGBA, offsetx int, offsety int, color color.RGBA, cellSizeW int, cellSizeH int) {
	for x := offsetx * cellSizeW; x < offsetx*cellSizeW+cellSizeW; x++ {
		for y := offsety * cellSizeH; y < offsety*cellSizeH+cellSizeH; y++ {
			m.Set(x, y, color)
		}
	}
}

func posUp(x, y int) (int, int, bool) {
	if y-1 < 0 {
		return x, y, false
	}
	return x, y - 1, true
}

func posDown(x, y, yLimit int) (int, int, bool) {
	if y+1 > yLimit {
		return x, y, false
	}
	return x, y + 1, true
}

func posLeft(x, y int) (int, int, bool) {
	if x-1 < 0 {
		return x, y, false
	}

	return x - 1, y, true
}

func posRight(x, y, xLimit int) (int, int, bool) {
	if x+1 > xLimit {
		return x, y, false
	}
	return x + 1, y, true
}

func DrawQRCode() *image.RGBA {
	cellNumber := 21 // 縦と横のそれぞれのセル数(21x21)
	w, h := 210, 210 // window size
	cellSizeW, cellSizeH := w/cellNumber, h/cellNumber
	m := image.NewRGBA(image.Rect(0, 0, w, h))

	white := color.RGBA{
		255,
		255,
		255,
		255,
	}
	black := color.RGBA{
		0,
		0,
		0,
		255,
	}

	for posx := 0; posx < cellNumber; posx++ {
		for posy := 0; posy < cellNumber; posy++ {
			if posx%2 == 0 && posy%2 == 0 {
				drawCell(m, posx, posy, white, cellSizeW, cellSizeH)
			} else {
				drawCell(m, posx, posy, black, cellSizeW, cellSizeH)
			}
		}
	}
	return m
}

func convertCharCode(i int) int {
	if 90 >= i && i >= 65 {
		return i - int('A') + 10
	}
	if 57 >= i && i >= 48 {
		return i - int('0')
	}
	return -1

}

func convertDataToBinaryString(str string) string {
	/*
		文字列を2進数になおす
		// 参考 http://c-faculty.chuo-u.ac.jp/~kuwata/2017-18/PDFs/Seminar_I_Wed/Student_ID-worksheet-b4.pdf
		英数字モードの場合は2文字で1組をあらわす．11ビットで表現(45進数)する必要がある
		文字列の長さが奇数の場合は最後の文字を6ビットで表現する
	*/
	result := ""
	for i := 0; i < len(str)/2; i++ {
		result += fmt.Sprintf("%011b", convertCharCode(int(str[i*2]))*45+convertCharCode(int(str[i*2+1])))
	}
	if len(str)%2 == 1 {
		result += fmt.Sprintf("%06b", convertCharCode(int(str[len(str)-1])))
	}
	return result
}

func charLength(str string, mode string) string {
	// 今は英数字モードのみ扱う
	return fmt.Sprintf("%09b", len(str))
}

func convertIntArray(str string) []uint {
	length := len(str) / 8
	padding := 8 - len(str)%8
	if padding != 8 {
		length += 1
	}
	var result []uint
	for i := 0; i < len(str)/8; i++ {
		parseUint, _ := strconv.ParseUint(str[i*8:i*8+8], 2, 8)
		result = append(result, uint(parseUint))
	}
	if padding != 8 {
		// 後半のビット+8ビットになるようにpadding(埋め合わせ)
		parseInt, _ := strconv.ParseUint(str[int(len(str)/8)*8:]+fmt.Sprintf("%0"+fmt.Sprint(padding)+"d", 0), 2, 8)
		result = append(result, uint(parseInt))
	}
	return result
}

func paddingDataCode(data []uint) []uint {
	// 1-H型は9コード(長さが9)である
	H1 := 9
	if len(data) < H1 {
		for i := 0; i < H1-len(data); i++ {
			var padding uint64
			if i%2 == 0 {
				padding, _ = strconv.ParseUint("11101100", 2, 8)
			} else {

				padding, _ = strconv.ParseUint("00010001", 2, 8)
			}
			data = append(data, uint(padding))
		}
	}
	return data
}

func fixedPatterns(bitmap [][]int, cellSize, fixedPos int) {
	for i := 0; i < cellSize; i++ {
		bitmap[fixedPos][i] = i % 2
		bitmap[i][fixedPos] = i % 2
	}
	finderPatternLeftUps := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 2},
		{0, 2, 2, 2, 2, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 2, 2, 2, 2, 0, 2},
		{0, 0, 0, 0, 0, 0, 0, 2},
		{2, 2, 2, 2, 2, 2, 2, 2},
	}
	finderPatternRightUps := [][]int{
		{2, 0, 0, 0, 0, 0, 0, 0},
		{2, 0, 2, 2, 2, 2, 2, 0},
		{2, 0, 2, 0, 0, 0, 2, 0},
		{2, 0, 2, 0, 0, 0, 2, 0},
		{2, 0, 2, 0, 0, 0, 2, 0},
		{2, 0, 2, 2, 2, 2, 2, 0},
		{2, 0, 0, 0, 0, 0, 0, 0},
		{2, 2, 2, 2, 2, 2, 2, 2},
	}
	finderPatternLeftDowns := [][]int{
		{2, 2, 2, 2, 2, 2, 2, 2},
		{0, 0, 0, 0, 0, 0, 0, 2},
		{0, 2, 2, 2, 2, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 0, 0, 0, 2, 0, 2},
		{0, 2, 2, 2, 2, 2, 0, 2},
		{0, 0, 0, 0, 0, 0, 0, 2},
	}
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			bitmap[x][y] = finderPatternLeftUps[y][x]
			bitmap[x+13][y] = finderPatternRightUps[y][x]
			bitmap[x][y+13] = finderPatternLeftDowns[y][x]
		}
	}
	for i := 0; i < 8; i++ {
		if i != fixedPos {
			bitmap[8][i] = 2
			bitmap[i][8] = 2
		}
		if i != fixedPos+13 {
			bitmap[i+13][8] = 2
			bitmap[8][i+13] = 2
		}

	}
	bitmap[8][8] = 2
	bitmap[13][8] = 0
}

func drawData(bitmap [][]int, bitString string) {

	pattern := "rightleft"
	currentPosCol := len(bitmap)
	currentPosRow := len(bitmap[0])
	count := 0

	for row := len(bitmap); row >= 0; row-- {
		for col := len(bitmap[row]); col >= 0; col-- {
			if pattern == "rightleft" {
				if count < len(bitString) {
					if bitmap[currentPosCol][currentPosRow-1] != 2 {
						currentPosRow -= 1
						bitmap[currentPosCol][currentPosRow] = int(bitString[count])
					}
				} else {
					if bitmap[currentPosCol][currentPosRow-1] != 2 {
						currentPosRow -= 1
						bitmap[currentPosCol][currentPosRow] = 0
					}
				}
				pattern = "upright"
			}
			if pattern == "upright" {
				if count < len(bitString) {
					if bitmap[currentPosCol-1][currentPosRow+1] != 2 {
						currentPosRow += 1
						currentPosCol -= 1
						bitmap[currentPosCol][currentPosRow] = int(bitString[count])
					}
				} else {
					if bitmap[currentPosCol-1][currentPosRow+1] != 2 {
						currentPosRow += 1
						currentPosCol -= 1
						bitmap[currentPosCol][currentPosRow] = 0
					}
				}
				pattern = "rightleft"
			}
		}
	}
}

func createBitmap(data []uint) [][]int {
	result := make([][]int, 21)
	for i := 0; i < 21; i++ {
		result[i] = make([]int, 21)
	}
	bitString := ""
	for i := 0; i < len(data); i++ {
		bitString += fmt.Sprintf("%08b", data[i])
	}
	fixedPatterns(result, len(result), 6)
	drawData(result, bitString)
	fmt.Println(bitString)
	return result
}

//func generateRS(data []uint) []uint {
//	/* 1-H型は誤り訂正コード語数が17なので
//	x17 +α43x16+α139x15+α206x14 +α78x13+α43x12+α239x11 +α123x10+α206x9+α214x8 +α147x7+α24x6+α99x5 +α150x4+α39x3+α243x2 +α163x+α136
//	をつかう
//	*/
//}

func main() {
	message := "ABCDE123"
	data := ""

	// モード指示子
	mode := "0010"
	data += mode

	// 文字数指示子
	data += charLength(message, mode)

	// データ変換
	data += convertDataToBinaryString(message)

	// 終端パターン
	data += "0000"

	fmt.Println(data)
	// 8bitごとの配列に変換
	arrayData := convertIntArray(data)

	// データコード数の補完
	arrayData = paddingDataCode(arrayData)
	fmt.Println(arrayData)

	f, _ := os.Create("test1.png")
	defer f.Close()
	rs := RS{}
	rs.InitTables(0x11d)
	for i := 0; i < len(arrayData); i++ {
		fmt.Print(fmt.Sprintf("%x ", arrayData[i]))
	}
	fmt.Println()
	rsEncode := rs.RsEncodeMsg(arrayData, 17)
	for i := 0; i < len(rsEncode); i++ {
		arrayData = append(arrayData, uint(rsEncode[i]))
	}
	img := DrawQRCode()
	png.Encode(f, img)
	bitmap := createBitmap(arrayData)
	for x := 0; x < 21; x++ {
		for y := 0; y < 21; y++ {
			fmt.Print(" ")
			if bitmap[x][y] == 0 {
				fmt.Print("□")
			} else {
				fmt.Print("■")
			}
		}
		fmt.Println()
	}
	fmt.Print(fmt.Sprintf("%x ", arrayData))
	fmt.Println()
}
