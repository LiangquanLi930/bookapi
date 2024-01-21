package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	_ "os"
	"os/signal"
	_ "path/filepath"
	"syscall"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
)

type Book struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var dataFilePath = "/data.json"

func main() {
	// 在启动服务器之前加载数据
	loadDataFromFile()

	// 创建一个新的 WebService
	ws := new(restful.WebService)
	ws.Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// 注册 API 路由
	ws.Route(ws.GET("/books").To(getBooks).
		Doc("获取所有 books").
		Writes([]Book{}))

	ws.Route(ws.GET("/books/{book-id}").To(getBook).
		Doc("通过 ID 获取 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")).
		Writes(Book{}))

	ws.Route(ws.POST("/books/").To(createBook).
		Doc("创建新的 book").
		Reads(Book{}))

	ws.Route(ws.PUT("/books/{book-id}").To(updateBook).
		Doc("更新指定 ID 的 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")).
		Reads(Book{}))

	ws.Route(ws.DELETE("/books/{book-id}").To(deleteBook).
		Doc("删除指定 ID 的 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")))

	ws.Route(ws.GET("/exit").To(exitHandler).
		Doc("退出程序"))

	// 将 WebService 注册到 Container
	restful.Add(ws)

	// 启动 HTTP 服务器
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	// 等待退出信号
	waitForExitSignal()
}

func exitHandler(request *restful.Request, response *restful.Response) {
	fmt.Println("Received exit request. Exiting...")
	os.Exit(0)
}

func waitForExitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("Received signal: %v. Exiting...\n", sig)
	os.Exit(0)
}

func getBooks(request *restful.Request, response *restful.Response) {
	// 返回所有 books
	books := loadDataFromFile()
	response.WriteEntity(books)
}

func getBook(request *restful.Request, response *restful.Response) {
	bookID := request.PathParameter("book-id")
	// 通过 ID 获取 book
	books := loadDataFromFile()
	if book, found := books[bookID]; found {
		response.WriteEntity(book)
	} else {
		response.WriteError(http.StatusNotFound, restful.NewError(http.StatusNotFound, "Book not found"))
	}
}

func createBook(request *restful.Request, response *restful.Response) {
	book := new(Book)
	err := request.ReadEntity(book)
	if err == nil {
		// 生成新的 ID
		book.ID = generateID()
		// 添加新的 book
		books := loadDataFromFile()
		books[book.ID] = *book
		saveDataToFile(books)
		response.WriteHeaderAndEntity(http.StatusCreated, book)
	} else {
		response.WriteError(http.StatusBadRequest, err)
	}
}

func updateBook(request *restful.Request, response *restful.Response) {
	bookID := request.PathParameter("book-id")
	books := loadDataFromFile()
	existingBook, found := books[bookID]
	if !found {
		response.WriteError(http.StatusNotFound, restful.NewError(http.StatusNotFound, "Book not found"))
		return
	}

	updatedBook := new(Book)
	err := request.ReadEntity(updatedBook)
	if err == nil {
		// 更新 book
		existingBook.Title = updatedBook.Title
		books[bookID] = existingBook
		saveDataToFile(books)
		response.WriteEntity(existingBook)
	} else {
		response.WriteError(http.StatusBadRequest, err)
	}
}

func deleteBook(request *restful.Request, response *restful.Response) {
	bookID := request.PathParameter("book-id")
	// 删除指定 ID 的 book
	books := loadDataFromFile()
	delete(books, bookID)
	saveDataToFile(books)
	response.WriteHeader(http.StatusNoContent)
}

func generateID() string {
	// 使用 UUID 生成唯一的 ID
	id := uuid.New().String()
	return "id" + id
}

func loadDataFromFile() map[string]Book {
	// 读取文件中的数据
	data, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		// 如果文件不存在，则返回空的 map
		return map[string]Book{}
	}

	// 解析 JSON 数据
	var books map[string]Book
	err = json.Unmarshal(data, &books)
	if err != nil {
		panic(err)
	}

	return books
}

func saveDataToFile(books map[string]Book) {
	// 将数据序列化为 JSON 格式
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		panic(err)
	}

	// 写入文件
	err = ioutil.WriteFile(dataFilePath, data, 0644)
	if err != nil {
		panic(err)
	}
}
