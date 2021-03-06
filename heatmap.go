package chart

import (
	"errors"
	"io"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/drawing"
)

// A Heatmap is a row of Histograms.
type Heatmap struct {
	Title     string
	Width     int
	Height    int
	DPI       float64
	Grid      [][]float64
	RowLabels []string
	ColLabels []string
}

type Column struct {
	Label  string
	Values []float64
}

type cell struct {
	Box
	Value float64
}

// Render renders the receiving Heatmap using the given RenderProvider or
// Writer.
func (h Heatmap) Render(rp RendererProvider, w io.Writer) error {
	if len(h.Grid) < 1 {
		return errors.New("Heatmap has no data to renderer")
	}

	columnLen := len(h.Grid[0])
	for _, column := range h.Grid {
		if len(column) != columnLen {
			return errors.New("Heatmap columns must all be the same length")
		}
	}

	if len(h.RowLabels) != len(h.Grid[0]) {
		return errors.New("Number of row labels != number of rows")
	}
	if len(h.ColLabels) != len(h.Grid) {
		return errors.New("Number of col lables != number of cols")
	}

	r, err := rp(h.Width, h.Height)
	if err != nil {
		return err
	}

	r.SetDPI(DefaultDPI)
	h.drawBackground(r)

	cellsBox := h.cellsBox()
	cells := h.computeCells(h.Grid, cellsBox)
	for _, cell := range cells {
		h.drawCell(r, cell)
	}

	Draw.Box(r, h.columnLabelBox(), Style{
		FillColor: drawing.ColorRed,
	})
	Draw.Box(r, h.rowLabelBox(), Style{
		FillColor: drawing.ColorGreen,
	})

	for col, label := range h.ColLabels {
		h.drawColumnLabel(r, label, cells[col*len(h.Grid[0])])
	}
	for row, label := range h.RowLabels {
		h.drawRowLabel(r, label, cells[row])
	}
	return r.Save(w)
}

func (h *Heatmap) drawBackground(r Renderer) {
	Draw.Box(r, h.box(),
		Style{
			FillColor:   drawing.ColorBlack,
			StrokeColor: drawing.ColorBlack,
			StrokeWidth: DefaultStrokeWidth,
		})
}

func (hm *Heatmap) computeCellSize(maxW int, maxH int) (w int, h int) {
	ncols := len(hm.Grid)
	nrows := len(hm.Grid[0])
	w = int(float64(maxW) / float64(ncols))
	h = int(float64(maxH) / float64(nrows))
	return
}

func (h *Heatmap) computeCells(grid [][]float64, box Box) []cell {
	cellWidth, cellHeight := h.computeCellSize(box.Width(), box.Height())
	var cells []cell
	for ci, column := range grid {
		for ri, value := range column {
			cells = append(cells, cell{
				Value: value,
				Box: Box{
					Top:    box.Top + ri*cellHeight,
					Bottom: box.Top + (ri+1)*cellHeight,
					Left:   box.Left + ci*cellWidth,
					Right:  box.Left + (ci+1)*cellWidth,
				},
			})
		}
	}
	return cells
}

func (h *Heatmap) drawCell(r Renderer, cell cell) {
	value := cell.Value
	box := cell.Box
	Draw.Box(r, box, Style{
		FillColor:   h.computeColor(value),
		StrokeColor: drawing.ColorBlack,
	})
}

func (h *Heatmap) drawColumnLabel(r Renderer, label string, topCell cell) {
	labelX := topCell.Box.Left + topCell.Box.Width()/2.0
	labelY := topCell.Box.Top - 10

	Draw.Text(r, label, labelX, labelY, Style{
		FontColor:           drawing.ColorBlack,
		FontSize:            18,
		Font:                h.font(),
		TextRotationDegrees: -90,
	})
}

func (h *Heatmap) drawRowLabel(r Renderer, label string, leftCell cell) {
	labelX := 0
	labelY := leftCell.Box.Top + leftCell.Box.Height()/2.0

	Draw.Text(r, label, labelX, labelY, Style{
		FontColor: drawing.ColorBlack,
		FontSize:  18,
		Font:      h.font(),
	})
}
func (h *Heatmap) font() *truetype.Font {
	// box := h.cellsBox()
	f, err := GetDefaultFont()
	if err != nil {
		panic(err)
	}
	return f
}

func (h *Heatmap) computeColor(value float64) drawing.Color {
	maxValue := h.maxValue()
	var r = 255 - uint32((value/maxValue)*255)
	var g = 255 - uint32((value/maxValue)*255)
	var b uint32 = 255
	return drawing.ColorFromAlphaMixedRGBA(r, g, b, 255)
}

// box returns the chart bounds as a box.
func (h *Heatmap) box() Box {
	return Box{
		Top:    0,
		Left:   0,
		Right:  h.Width,
		Bottom: h.Height,
	}
}

func (h *Heatmap) cellsBox() Box {
	clb := h.columnLabelBox()
	rlb := h.rowLabelBox()

	return Box{
		Top:    clb.Bottom,
		Left:   rlb.Right,
		Right:  h.Width,
		Bottom: h.Height,
	}
}

func (h *Heatmap) columnLabelBox() Box {
	box := h.box()
	rlb := h.rowLabelBox()

	return Box{
		Top:    0,
		Left:   rlb.Right,
		Right:  box.Right,
		Bottom: 300,
	}
}

func (h *Heatmap) rowLabelBox() Box {
	return Box{
		Top:    0,
		Left:   0,
		Right:  300,
		Bottom: h.Height,
	}
}

func (h *Heatmap) maxValue() float64 {
	maxValue := h.Grid[0][0]
	for _, col := range h.Grid {
		for _, value := range col {
			if value > maxValue {
				maxValue = value
			}
		}
	}
	return maxValue
}
