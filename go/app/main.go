package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	//"crypto/sha256"
	//"encoding/hex"
	"encoding/json"
	//"errors"
	//"io"
	//"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Items struct {
	Items []Item `json:"item"`
}

type Item struct {
	//ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	//ImageName string `json:"image_name"`
}

const (
	ImgDir = "images"
	JSONFile = "data/items.json"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}


func readItems() (*Items, error) {
	jsonItemData, err := os.ReadFile(JSONFile)
	if err != nil {
		return nil, err
	}
	var addItems Items
	// Decode: JSONからItemsに変換
	if err := json.Unmarshal(jsonItemData, &addItems); err != nil {
		return nil, err
	}
	return &addItems, nil
}

// ItemsからJSONに変換
func writeItems(items *Items) error {
	fmt.Printf("Writing items to file: %+v\n", items.Items) // ログ出力: アイテムをファイルに書き込む前にアイテムを出力
	jsonItemData, err := os.Create(JSONFile)
	if err != nil {
		return err
	}
	defer jsonItemData.Close()
	// Encode: ItemsからJSONに変換
	encoder := json.NewEncoder(jsonItemData)
	if err := encoder.Encode(items); err != nil {
		return err
	}
	fmt.Printf("Items written to file: %s\n", JSONFile)
	return nil
}



func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	
	newItem := Item{Name: name, Category: category}

	// Read existing items from JSON file
	items, err := readItems()
	if err != nil {
		c.Logger().Errorf("Failed to read items: %v", err)
		return err
	}

	// Append new item
	items.Items = append(items.Items, newItem)

	// Write items to JSON file
	if err := writeItems(items); err != nil {
		c.Logger().Errorf("Failed to write items: %v", err)
		return err
	}

	message := fmt.Sprintf("Item received: %s, category: %s", newItem.Name, newItem.Category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}
func getItems(c echo.Context) error {
	items, err := readItems()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func printItemsJSON() error {
	jsonItemData, err := os.ReadFile(JSONFile)
	if err != nil {
		return err
	}

	fmt.Println("Current items.json contents:")
	fmt.Println(string(jsonItemData))
	return nil
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.DEBUG)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	if err := printItemsJSON(); err != nil {
        fmt.Printf("Failed to print items.json contents: %v\n", err)
    }

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
