package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	//"errors"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

// ItemsがItem構造体への、値のスライス→ポインタのスライスに変更
type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	//ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

const (
	ImgDir   = "images"
	JSONFile = "items.json"
	dbPath   = "/Users/miwa/mercari/mercari-build-training/go/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		res := Response{Message: "Return image FormFile"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
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

	// データベースへの接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		res := Response{Message: "Error opening file"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	defer db.Close()

	// カテゴリが存在するか調べる
	var categoryID int64
	row := db.QueryRow("SELECT id FROM categories WHERE name = ?", category)
	err = row.Scan(&categoryID)
	// カテゴリが存在しない場合、新しいカテゴリを追加
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			res := Response{Message: "Error adding new categories to the database"}
			return echo.NewHTTPError(http.StatusInternalServerError, res)
		}
		categoryID, _ = result.LastInsertId()
	} else if err != nil {
		c.Logger().Errorf("Error INSERT INTO items: %s", err)
		res := Response{Message: "Error querying categories from the database"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	// アイテムをデータベースに追加
	stmt, err := db.Prepare("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)")
	if err != nil {
		c.Logger().Errorf("Error preparing SQL statement: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error preparing SQL statement")
	}
	defer stmt.Close()

	if _, err = stmt.Exec(name, category, imageHash+".jpg"); err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		res := Response{Message: "Error opening file"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	message := fmt.Sprintf("Item received: %s, category: %s, image_name: %s", name, category, imageHash+".jpg")
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
	// データベースへの接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Logger().Errorf("Error opening database: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error opening file")
	}
	defer db.Close()

	// SQLクエリの実行
	rows, err := db.Query("SELECT * FROM items;")
	if err != nil {
		c.Logger().Errorf("Error querying items: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}
	defer rows.Close()

	//items := new(Items)
	var items []Item

	// レコードをスキャンしてItem構造体に変換
	for rows.Next() {
		var item Item
		// レコードの各カラムをItem構造体にスキャン
		if err := rows.Scan(&item.Name, &item.Category, &item.ImageName); err != nil {
			c.Logger().Errorf("Error scanning item: %s", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error scanning item")
		}
		items = append(items, item)
	}
	// エラーがあればログ出力
	if err := rows.Err(); err != nil {
		c.Logger().Errorf("Error iterating over rows: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error iterating over rows")
	}

	//json形式に変換
	return c.JSON(http.StatusOK, map[string][]Item{"items": items})
}

func getItemsId(c echo.Context) error {
	// データベースへの接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error opening file")
	}
	defer db.Close()

	//idを取得
	id := c.Param("id")
	itemID, err := strconv.Atoi(id)
	if err != nil {
		res := Response{Message: "Error geting itemID"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	// 指定されたIDのアイテムを取得
	var item Item
	query := "SELECT items.name, categories.name as categories, items.image_name FROM items join categories on items.category_id = categories.id WHERE items.id = ?"
	row := db.QueryRow(query, itemID)
	err = row.Scan(&item.Name, &item.Category, &item.ImageName)
	if err != nil {
		c.Logger().Errorf("Error Query: %s", err)
		res := Response{Message: "Error Query"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, item)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Error image path"}
		return echo.NewHTTPError(http.StatusInternalServerError, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func searchItem(c echo.Context) error {
	var items Items
	//keyword := c.FormValue("keyword")
	keyword := c.QueryParam("keyword")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	cmd := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?"
	rows, err := db.Query(cmd, "%"+keyword+"%")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, category, imageName string
		if err := rows.Scan(&name, &category, &imageName); err != nil {
			return err
		}
		item := Item{Name: name, Category: category, ImageName: imageName}
		items.Items = append(items.Items, item)
	}
	return c.JSON(http.StatusOK, items)
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

	// Routes
	e.GET("/", root)
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/items/:id", getItemsId)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", searchItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
