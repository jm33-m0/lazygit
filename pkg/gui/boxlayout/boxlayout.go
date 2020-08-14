package boxlayout

import "math"

type Dimensions struct {
	X0 int
	X1 int
	Y0 int
	Y1 int
}

const (
	ROW = iota
	COLUMN
)

// to give a high-level explanation of what's going on here. We layout our views by arranging a bunch of boxes in the window.
// If a box has children, it needs to specify how it wants to arrange those children: ROW or COLUMN.
// If a box represents a view, you can put the view name in the viewName field.
// When determining how to divvy-up the available height (for row children) or width (for column children), we first
// give the boxes with a static `size` the space that they want. Then we apportion
// the remaining space based on the weights of the dynamic boxes (you can't define
// both size and weight at the same time: you gotta pick one). If there are two
// boxes, one with weight 1 and the other with weight 2, the first one gets 33%
// of the available space and the second one gets the remaining 66%

type Box struct {
	// Direction decides how the children boxes are laid out. ROW means the children will each form a row i.e. that they will be stacked on top of eachother.
	Direction int // ROW or COLUMN

	// function which takes the width and height assigned to the box and decides which orientation it will have
	ConditionalDirection func(width int, height int) int

	Children []*Box

	// function which takes the width and height assigned to the box and decides the layout of the children.
	ConditionalChildren func(width int, height int) []*Box

	// ViewName refers to the name of the view this box represents, if there is one
	ViewName string

	// static Size. If parent box's direction is ROW this refers to height, otherwise width
	Size int

	// dynamic size. Once all statically sized children have been considered, Weight decides how much of the remaining space will be taken up by the box
	// TODO: consider making there be one int and a type enum so we can't have size and Weight simultaneously defined
	Weight int
}

func ArrangeViews(root *Box, x0, y0, width, height int) map[string]Dimensions {
	children := root.getChildren(width, height)
	if len(children) == 0 {
		// leaf node
		if root.ViewName != "" {
			dimensionsForView := Dimensions{X0: x0, Y0: y0, X1: x0 + width - 1, Y1: y0 + height - 1}
			return map[string]Dimensions{root.ViewName: dimensionsForView}
		}
		return map[string]Dimensions{}
	}

	direction := root.getDirection(width, height)

	var availableSize int
	if direction == COLUMN {
		availableSize = width
	} else {
		availableSize = height
	}

	// work out size taken up by children
	reservedSize := 0
	totalWeight := 0
	for _, child := range children {
		// assuming either size or weight are non-zero
		reservedSize += child.Size
		totalWeight += child.Weight
	}

	remainingSize := availableSize - reservedSize
	if remainingSize < 0 {
		remainingSize = 0
	}

	unitSize := 0
	extraSize := 0
	if totalWeight > 0 {
		unitSize = remainingSize / totalWeight
		extraSize = remainingSize % totalWeight
	}

	result := map[string]Dimensions{}
	offset := 0
	for _, child := range children {
		var boxSize int
		if child.isStatic() {
			boxSize = child.Size
		} else {
			// TODO: consider more evenly distributing the remainder
			boxSize = unitSize * child.Weight
			boxExtraSize := int(math.Min(float64(extraSize), float64(child.Weight)))
			boxSize += boxExtraSize
			extraSize -= boxExtraSize
		}

		var resultForChild map[string]Dimensions
		if direction == COLUMN {
			resultForChild = ArrangeViews(child, x0+offset, y0, boxSize, height)
		} else {
			resultForChild = ArrangeViews(child, x0, y0+offset, width, boxSize)
		}

		result = mergeDimensionMaps(result, resultForChild)
		offset += boxSize
	}

	return result
}

func (b *Box) isStatic() bool {
	return b.Size > 0
}

func (b *Box) getDirection(width int, height int) int {
	if b.ConditionalDirection != nil {
		return b.ConditionalDirection(width, height)
	}
	return b.Direction
}

func (b *Box) getChildren(width int, height int) []*Box {
	if b.ConditionalChildren != nil {
		return b.ConditionalChildren(width, height)
	}
	return b.Children
}

func mergeDimensionMaps(a map[string]Dimensions, b map[string]Dimensions) map[string]Dimensions {
	result := map[string]Dimensions{}
	for _, dimensionMap := range []map[string]Dimensions{a, b} {
		for k, v := range dimensionMap {
			result[k] = v
		}
	}
	return result
}
