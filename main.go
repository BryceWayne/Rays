package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	WIDTH     int = 256 * 4      // Screen width
	HEIGHT    int = 256 * 4      // Screen height
	randIndex int = rand.Intn(8) // Randomly select a tile index (0-7 inclusive
	startX    int = 4
	startY    int = 4
)

type Tile struct {
	Image          *ebiten.Image
	PossibleTop    []int
	PossibleBottom []int
	PossibleLeft   []int
	PossibleRight  []int
}

var tiles [8]Tile

var tileConnections = map[int]map[string][]int{
	0: {
		"top":    []int{0, 6},
		"bottom": []int{0, 5},
		"left":   []int{2, 4, 7},
		"right":  []int{2},
	},
	1: {
		"top":    []int{3, 4, 5},
		"bottom": []int{3},
		"left":   []int{1, 6},
		"right":  []int{1, 7},
	},
	2: {
		"top":    []int{2, 7},
		"bottom": []int{2, 4},
		"left":   []int{0},
		"right":  []int{0, 5, 6},
	},
	3: {
		"top":    []int{1},
		"bottom": []int{1, 6, 7},
		"left":   []int{3, 5},
		"right":  []int{3, 4},
	},
	4: {
		"top":    []int{2, 7},
		"bottom": []int{1, 6, 7},
		"left":   []int{3, 5},
		"right":  []int{0, 5, 6},
	},
	5: {
		"top":    []int{0, 6},
		"bottom": []int{1, 6, 7},
		"left":   []int{2, 4, 7},
		"right":  []int{3, 4},
	},
	6: {
		"top":    []int{3, 4, 5},
		"bottom": []int{0, 5},
		"left":   []int{2, 4, 7},
		"right":  []int{1, 7},
	},
	7: {
		"top":    []int{1},
		"bottom": []int{2, 4},
		"left":   []int{1, 6},
		"right":  []int{0, 5, 6},
	},
}

func initTiles() {
	for i := 0; i < 8; i++ {
		img, _, err := ebitenutil.NewImageFromFile(fmt.Sprintf("images/tilesets/desert/tile%d.png", i))
		if err != nil {
			log.Fatalf("Failed to load tile image: %v", err)
		}
		tiles[i] = Tile{
			Image:          img,
			PossibleTop:    tileConnections[i]["top"],
			PossibleBottom: tileConnections[i]["bottom"],
			PossibleLeft:   tileConnections[i]["left"],
			PossibleRight:  tileConnections[i]["right"],
		}
	}
}

func (t *Tile) Entropy() int {
	return len(t.PossibleTop) + len(t.PossibleBottom) + len(t.PossibleLeft) + len(t.PossibleRight)
}

type Cell struct {
	PossibleTiles []int // Indices of possible tiles
	Entropy       int   // Number of possible tiles
}

func initializeGrid(width, height int) [][]Cell {
	grid := make([][]Cell, height)
	for y := range grid {
		grid[y] = make([]Cell, width)
		for x := range grid[y] {
			// Initialize each cell with all possible tile indices
			grid[y][x].PossibleTiles = make([]int, len(tiles))
			for i := range tiles {
				grid[y][x].PossibleTiles[i] = i
			}
			// Set the entropy to the count of possible tiles
			grid[y][x].Entropy = len(grid[y][x].PossibleTiles)
		}
	}
	return grid
}

func init() {
	// Initialize random seed
	rand.Seed(42)

	// // Initialize off-screen image
	// initTiles()
}

type Game struct {
	Grid [][]Cell
}

func collapseCell(cell *Cell) {
	if cell.Entropy > 1 { // Check if cell has more than one possible tile
		selectedIndex := rand.Intn(cell.Entropy)
		selectedTile := cell.PossibleTiles[selectedIndex]
		cell.PossibleTiles = []int{selectedTile} // Collapse to the selected tile
		cell.Entropy = 1                         // Update entropy after collapse
	}
}

func isTileCompatible(tileIndex, x, y, dx, dy int, grid [][]Cell) bool {
	// Check for negative indices or out-of-bound access
	if x-dx < 0 || y-dy < 0 || y-dy >= len(grid) || x-dx >= len(grid[y-dy]) {
		return false
	}

	neighborCell := grid[y][x]

	// Ensure the neighbor cell is collapsed to a single tile
	if len(neighborCell.PossibleTiles) != 1 {
		return false
	}

	collapsedTileIndex := neighborCell.PossibleTiles[0] // The single tile in the collapsed neighbor cell

	// Retrieve the connection rules for the collapsed tile
	connections := tileConnections[collapsedTileIndex]

	var direction string
	if dx == 1 {
		direction = "left"
	} else if dx == -1 {
		direction = "right"
	} else if dy == 1 {
		direction = "top"
	} else if dy == -1 {
		direction = "bottom"
	}

	// Check if the tileIndex is in the list of compatible tiles for the direction
	for _, compatibleTileIndex := range connections[direction] {
		if tileIndex == compatibleTileIndex {
			return true
		}
	}

	return false
}

func updateNeighbors(x, y int, grid [][]Cell) {
	// Define the relative positions of the neighbors
	neighborPositions := []struct{ dx, dy int }{
		{0, -1}, // Top
		{0, 1},  // Bottom
		{-1, 0}, // Left
		{1, 0},  // Right
	}

	log.Printf("Current cell: (%d, %d) - %+v \n", x, y, grid[y][x])

	// Iterate over each neighbor
	for _, pos := range neighborPositions {
		nx, ny := x+pos.dx, y+pos.dy

		// Check if neighbor is within grid bounds
		if nx >= 0 && ny >= 0 && ny < len(grid) && nx < len(grid[ny]) {
			neighbor := &grid[ny][nx]

			log.Printf("Before update: Cell at (%d, %d) has PossibleTiles: %v", nx, ny, neighbor.PossibleTiles)

			// Only update neighbors that have not been collapsed to a single tile
			if len(neighbor.PossibleTiles) > 1 {
				newPossibleTiles := []int{}

				// Check each possible tile for the neighbor
				for _, tileIndex := range neighbor.PossibleTiles {
					if isTileCompatible(tileIndex, x, y, pos.dx, pos.dy, grid) {
						newPossibleTiles = append(newPossibleTiles, tileIndex)
					}
				}

				// After updating PossibleTiles for a neighbor
				neighbor.PossibleTiles = newPossibleTiles
				neighbor.Entropy = len(newPossibleTiles) // Update Entropy to match the new PossibleTiles count
				log.Printf("After update: Cell at (%d, %d) has PossibleTiles: %v", nx, ny, neighbor.PossibleTiles)
			}
		}
	}
}

func (g *Game) Update() error {
	allCollapsed := true         // Assume all cells are collapsed initially
	minEntropy := len(tiles) + 1 // Start higher than the maximum possible entropy
	var minCell *Cell
	var minX, minY int

	// Iterate over the grid to find the cell with the minimum entropy
	for y, row := range g.Grid {
		for x, cell := range row {
			if x == startX && y == startY {
				continue
			}
			if len(cell.PossibleTiles) == 0 {
				log.Printf("Cell: %+v \n", cell)
				log.Printf("Contradiction found at cell: (%d, %d)", x, y)

				// Restart the WFC process
				g.Grid = initializeGrid(WIDTH/16, HEIGHT/16) // Assuming each tile is 16x16
				startX, startY = rand.Intn(WIDTH/16), rand.Intn(HEIGHT/16)
				collapseCellAtStart(startX, startY, &g.Grid)
				return nil // Exit the current Update call
			}

			if len(cell.PossibleTiles) > 1 {
				allCollapsed = false // Found a cell that is not collapsed, so not all cells are collapsed
				entropy := len(cell.PossibleTiles)
				if entropy < minEntropy { // Cell has not collapsed and has fewer possibilities
					minEntropy = entropy
					minCell = &g.Grid[y][x]
					minX, minY = x, y
				}
			}
		}
	}

	// If all cells are collapsed, the WFC process is complete
	if allCollapsed {
		log.Println("All cells collapsed. WFC process complete.")
		return nil // Return from the Update method to stop further updates
	}

	// Collapse the cell with the minimum entropy and update its neighbors
	if minCell != nil {
		collapseCell(minCell)
		updateNeighbors(minX, minY, g.Grid)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Modify this to draw the current state of each cell in the grid
	// For cells with multiple possible tiles, you might choose to draw nothing or a placeholder
	for y, row := range g.Grid {
		for x, cell := range row {
			if len(cell.PossibleTiles) == 1 {
				tileIndex := cell.PossibleTiles[0]
				tile := tiles[tileIndex]
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(x*16), float64(y*16)) // Assuming each tile is 16x16
				screen.DrawImage(tile.Image, opts)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Return the game's screen size
	return WIDTH, HEIGHT
}

func collapseCellAtStart(startX, startY int, grid *[][]Cell) {
	(*grid)[startY][startX].PossibleTiles = []int{randIndex} // Collapse cell at (0, 0) to the randomly selected tile
	(*grid)[startY][startX].Entropy = 1                      // Update entropy if you are tracking it
	updateNeighbors(startX, startY, *grid)
}

func main() {
	initTiles() // Ensure tiles are initialized before setting up the grid

	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetWindowTitle("My Dungeon Game")

	grid := initializeGrid(WIDTH/16, HEIGHT/16) // Assuming each tile is 16x16

	collapseCellAtStart(startX, startY, &grid)

	game := &Game{
		Grid: grid,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
