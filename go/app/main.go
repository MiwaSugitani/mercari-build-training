package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	//"errors"
	"io"
	"strconv"

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
	ImageName string `json:"image_name"`
}

const (
	ImgDir   = "images"
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
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		c.Logger().Errorf("Failed to get image: %v", err)
		return err
	}

	// 一時ファイルのパスを取得
	tempFilePath := path.Join(ImgDir, image.Filename)

	// アップロードされた画像を一時ファイルパスに保存
	if err := saveImage(image, tempFilePath); err != nil {
		c.Logger().Errorf("画像の保存に失敗しました: %v", err)
		return err
	}

	c.Logger().Infof("Image saved to: %s", tempFilePath)

	// 画像のハッシュ化
	imageFile, err := image.Open()
	if err != nil {
		c.Logger().Errorf("Failed to open image: %v", err)
		return err
	}
	defer imageFile.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, imageFile); err != nil {
		c.Logger().Errorf("Failed to hash image: %v", err)
		return err
	}
	imageHash := hex.EncodeToString(hash.Sum(nil))
	c.Logger().Infof("Image hash: %s", imageHash)

	// 画像の保存
	imagePath := path.Join(ImgDir, imageHash+".jpg")
	if err := saveImage(image, imagePath); err != nil {
		return err
	}
	c.Logger().Infof("Image saved to: %s", imagePath)
	// 新しいアイテムを作成
	newItem := Item{Name: name, Category: category, ImageName: imageHash + ".jpg"}

	// JSONファイルから既存のアイテムを読み取る
	items, err := readItems()
	if err != nil {
		c.Logger().Errorf("Failed to read items: %v", err)
		return err
	}

	// 新しいアイテムを追加
	items.Items = append(items.Items, newItem)
	c.Logger().Infof("New item added: %+v", newItem)

	// アイテムをJSONファイルに書き込む
	if err := writeItems(items); err != nil {
		c.Logger().Errorf("Failed to write items: %v", err)
		return err
	}

	message := fmt.Sprintf("Item received: %s, category: %s, image_name: %s", newItem.Name, newItem.Category, newItem.ImageName)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func saveImage(file *multipart.FileHeader, path string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func getItems(c echo.Context) error {
	items, err := readItems()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, items)
}

func getItemsId(c echo.Context) error {
	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Logger().Errorf("Invalid ID: %v", err)
		res := Response{Message: "Invalid ID"}
		return echo.NewHTTPError(http.StatusBadRequest, res)
	}

	items, err := readItems()
	if err != nil {
		c.Logger().Errorf("Error while reading item information: %v", err)
		res := Response{Message: "Error while reading item information"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	// アイテムのインデックスを取得
	index := itemId - 1
	if index < 0 || index >= len(items.Items) {
		c.Logger().Errorf("Invalid ID: %d", itemId)
		res := Response{Message: "Item not found"}
		return echo.NewHTTPError(http.StatusNotFound, res)
	}

	// 指定されたIDのアイテムを返す
	item := items.Items[index]
	return c.JSON(http.StatusOK, item)
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
	e.GET("/items/:id", getItemsId)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
