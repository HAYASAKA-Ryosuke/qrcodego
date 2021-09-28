package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
)

func drawCell(m *image.RGBA, row int, col int, color color.RGBA, cellSizeW int, cellSizeH int) {
	for x := row * cellSizeW; x < row*cellSizeW+cellSizeW; x++ {
		for y := col * cellSizeH; y < col*cellSizeH+cellSizeH; y++ {
			m.Set(y, x, color)
		}
	}
}

func DrawQRCode(bitmap [][]int) *image.RGBA {
	cellNumber := len(bitmap) // 縦と横のそれぞれのセル数(21x21)
	w, h := 210, 210          // window size
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

	for row := 0; row < cellNumber; row++ {
		for col := 0; col < cellNumber; col++ {
			if bitmap[row][col] == 0 {
				drawCell(m, row, col, black, cellSizeW, cellSizeH)
			} else {
				drawCell(m, row, col, white, cellSizeW, cellSizeH)
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
		{0, 0, 0, 0, 0, 0, 0, 1},
		{0, 1, 1, 1, 1, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 1, 1, 1, 1, 0, 1},
		{0, 0, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1, 1},
	}
	finderPatternRightUps := [][]int{
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 0, 1, 1, 1, 1, 1, 0},
		{1, 0, 1, 0, 0, 0, 1, 0},
		{1, 0, 1, 0, 0, 0, 1, 0},
		{1, 0, 1, 0, 0, 0, 1, 0},
		{1, 0, 1, 1, 1, 1, 1, 0},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 1, 1, 1},
	}
	finderPatternLeftDowns := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 0, 0, 0, 0, 1},
		{0, 1, 1, 1, 1, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 0, 0, 0, 1, 0, 1},
		{0, 1, 1, 1, 1, 1, 0, 1},
		{0, 0, 0, 0, 0, 0, 0, 1},
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
			bitmap[8][i] = 1
			bitmap[i][8] = 1
		}
		if i != fixedPos+13 {
			bitmap[i+13][8] = 1
			bitmap[8][i+13] = 1
		}

	}
	bitmap[8][8] = 1
	bitmap[13][8] = 0
}

func moveLeft(row, col int) (int, int) {
	return row, col - 1
}

func moveRight(row, col int) (int, int) {
	return row, col + 1
}

func moveDown(row, col int) (int, int) {
	return row + 1, col
}
func moveUp(row, col int) (int, int) {
	return row - 1, col
}

func moveUpRight(row, col int) (int, int) {
	return row - 1, col + 1
}
func moveDownRight(row, col int) (int, int) {
	return row + 1, col + 1
}

func drawData(bitmap [][]int, bitString string) {
	currentPosRow := len(bitmap) - 1
	currentPosCol := len(bitmap[0]) - 1
	count := 0
	for {
		if currentPosCol != 6 && currentPosRow != 6 {
			if count < len(bitString) {
				fmt.Println(int(bitString[count]) - 48)
				if int(bitString[count]) == '0' {
					bitmap[currentPosRow][currentPosCol] = 1
				} else {
					bitmap[currentPosRow][currentPosCol] = 0
				}
				count++
				drawMaskPattern(bitmap, currentPosRow, currentPosCol)
			} else {
				bitmap[currentPosRow][currentPosCol] = 0
			}
		}
		fmt.Printf("Row: %d, Col: %d\n", currentPosRow, currentPosCol)
		if currentPosCol == 0 && currentPosRow == 12 {
			break
		} else if currentPosCol == 6 {
			currentPosCol -= 1
		} else if currentPosCol < 6 {
			if currentPosCol%4 == 0 {
				if currentPosRow == 12 {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveDownRight(currentPosRow, currentPosCol)
				}
			} else if currentPosCol%4 == 1 {
				currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
			} else if currentPosCol%4 == 2 {
				if currentPosRow == 9 {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveUpRight(currentPosRow, currentPosCol)
				}
			} else if currentPosCol%4 == 3 {
				currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
			}
		} else {
			if (currentPosCol-7)%4 == 0 {
				if currentPosRow == 0 || (currentPosRow == 9 && (currentPosCol <= 8 || currentPosCol >= 13)) {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveUpRight(currentPosRow, currentPosCol)
				}
			} else if (currentPosCol-7)%4 == 1 {
				if currentPosRow == 6 {
					currentPosRow, currentPosCol = moveUp(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				}
			} else if (currentPosCol-7)%4 == 2 {
				if currentPosRow == 20 && currentPosCol == 9 {
					// ここは12,8に飛ぶ必要がある
					currentPosCol = 8
					currentPosRow = 12
				} else if currentPosRow == 20 {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveDownRight(currentPosRow, currentPosCol)
				}
			} else if (currentPosCol-7)%4 == 3 {
				if currentPosRow == 6 {
					currentPosRow, currentPosCol = moveDown(currentPosRow, currentPosCol)
				} else {
					currentPosRow, currentPosCol = moveLeft(currentPosRow, currentPosCol)
				}
			}
		}
	}
}

func drawMaskPattern(bitmap [][]int, row, col int) {
	// 000
	if (row+col)%3 == 0 {
		if bitmap[row][col] == 0 {
			bitmap[row][col] = 1
		} else {
			bitmap[row][col] = 0
		}
	}

}

func drawFormatbit(bitmap [][]int) {
	bitString := "001100111010000"
	count := 0
	for col := 0; col < 8; col++ {
		if col != 6 {
			if bitString[count] == '0' {
				bitmap[8][col] = 1
			} else {
				bitmap[8][col] = 0
			}
			count++
		}
	}
	for row := 0; row < 9; row++ {
		if row != 6 {
			if bitString[count] == '0' {
				bitmap[row][8] = 1
			} else {
				bitmap[row][8] = 0
			}
			count++
		}
	}
	count = 0
	for row := 0; row < 7; row++ {
		if bitString[count] == '0' {
			bitmap[20-row][8] = 1
		} else {
			bitmap[20-row][8] = 0
		}
		count++
	}
	for col := 0; col < 8; col++ {
		if bitString[count] == '0' {
			bitmap[8][col+13] = 1
		} else {
			bitmap[8][col+13] = 0
		}
		count++
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
	drawFormatbit(result)
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
	fmt.Println(bitmap[19][20])
	img := DrawQRCode(bitmap)
	png.Encode(f, img)
}
