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

func NewImage() *image.RGBA {
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
	img := NewImage()
	png.Encode(f, img)
}
