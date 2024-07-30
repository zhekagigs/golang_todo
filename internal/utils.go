package internal

import (
	"embed"
	"fmt"
	"io/fs"
	"math/rand"
	"strings"
	"time"
)

//go:embed resources
var resourcesFiles embed.FS

var timeNow = time.Now

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func formatDatetime(t time.Time) string {
	return t.Format("Monday, January 2, 2006 at 15:04")
}

func generateRandomTasks(count int) []Task {
	brewingVerbs := []string{"Brew", "Ferment", "Bottle", "Label", "Clean", "Inspect", "Order", "Taste"}
	adjectives := []string{"Hoppy", "Malty", "Crisp", "Smooth", "Tangy", "Barrel-aged", "Experimental", "Funky", "Juicy", "Hazy", "Robust", "Refreshing"}
	beerTypes := []string{"IPA", "Stout", "Lager", "Wheat Beer", "Pale Ale", "Porter", "Sour", "Seasonal Batch"}
	ingredients := []string{"Hops", "Malt", "Yeast", "Fruit puree", "Spices", "Coffee beans", "Cocoa nibs", "Oak chips"}
	marketingTasks := []string{"Design new label", "Plan social media campaign", "Organize tasting event", "Create promotional video", "Update website", "Develop brand partnership", "Analyze market trends", "Conduct customer survey"}
	logisticsTasks := []string{"Schedule delivery route", "Inventory check", "Restock supplies", "Maintain delivery vehicles", "Optimize warehouse layout", "Negotiate with suppliers", "Update inventory management system", "Coordinate with distributors"}
	qualityTasks := []string{"Conduct sensory analysis", "Calibrate testing equipment", "Review quality control procedures", "Train staff on quality standards", "Perform microbiological testing", "Update quality assurance documentation", "Conduct supplier quality audit", "Implement new quality control measure"}

	tasks := make([]Task, count)
	for i := 0; i < count; i++ {
		var taskValue string
		var category TaskCategory
		x := rand.Intn(5)
		switch x {
		case 0:
			verb := brewingVerbs[rand.Intn(len(brewingVerbs))]
			adj := adjectives[rand.Intn(len(adjectives))]
			noun := beerTypes[rand.Intn(len(beerTypes))]
			taskValue = strings.Join([]string{verb, adj, noun}, " ")
			category = Brewing
		case 1:
			verb := "Process"
			if rand.Intn(2) == 0 {
				verb = "Order"
			}
			ingredient := ingredients[rand.Intn(len(ingredients))]
			quantity := rand.Intn(500) + 1 // 1 to 500
			unit := "kg"
			if rand.Intn(2) == 0 {
				unit = "lbs"
			}
			taskValue = fmt.Sprintf("%s %d %s of %s", verb, quantity, unit, ingredient)
			category = Brewing
		case 2:
			taskValue = marketingTasks[rand.Intn(len(marketingTasks))]
			category = Marketing
		case 3:
			taskValue = logisticsTasks[rand.Intn(len(logisticsTasks))]
			category = Logistics
		case 4:
			taskValue = qualityTasks[rand.Intn(len(qualityTasks))]
			category = Quality

		}
		plannedAt := time.Now().Add(time.Duration(rand.Intn(24*30)) * time.Hour)
		newTask := NewTask(i, taskValue, category, plannedAt)
		tasks[i] = newTask
	}
	return tasks
}

func BeerAscii() string {
	data, err := fs.ReadFile(resourcesFiles, "resources/beer.txt")
	if err != nil {
		fmt.Println("error happened while reading beer logo: ", err)
		return ""
	}
	return string(data)
}

func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func CategoryPtr(c TaskCategory) *TaskCategory {
	return &c
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func BoolPtr(b bool) *bool {
	return &b
}
