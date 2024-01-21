package main

import (
	"github.com/google/uuid"
	"net/http"

	"github.com/emicklei/go-restful"
)

type Book struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var books = map[string]Book{}

func main() {
	// 创建一个新的 WebService
	ws := new(restful.WebService)
	ws.Path("/books").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// 注册 API 路由
	ws.Route(ws.GET("/").To(getItems).
		Doc("获取所有 books").
		Writes([]Book{}))

	ws.Route(ws.GET("/{book-id}").To(getItem).
		Doc("通过 ID 获取 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")).
		Writes(Book{}))

	ws.Route(ws.POST("/").To(createItem).
		Doc("创建新的 book").
		Reads(Book{}))

	ws.Route(ws.PUT("/{book-id}").To(updateItem).
		Doc("更新指定 ID 的 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")).
		Reads(Book{}))

	ws.Route(ws.DELETE("/{book-id}").To(deleteItem).
		Doc("删除指定 ID 的 book").
		Param(ws.PathParameter("book-id", "book 的 ID").DataType("string")))

	// 将 WebService 注册到 Container
	restful.Add(ws)

	// 启动 HTTP 服务器
	http.ListenAndServe(":8080", nil)
}

func getItems(request *restful.Request, response *restful.Response) {
	// 返回所有 items
	response.WriteEntity(books)
}

func getItem(request *restful.Request, response *restful.Response) {
	itemID := request.PathParameter("book-id")
	// 通过 ID 获取 book
	if book, found := books[itemID]; found {
		response.WriteEntity(book)
	} else {
		response.WriteError(http.StatusNotFound, restful.NewError(http.StatusNotFound, "book not found"))
	}
}

func createItem(request *restful.Request, response *restful.Response) {
	book := new(Book)
	err := request.ReadEntity(book)
	if err == nil {
		// 生成新的 ID
		book.ID = generateID()
		// 添加新的 book
		books[book.ID] = *book
		response.WriteHeaderAndEntity(http.StatusCreated, book)
	} else {
		response.WriteError(http.StatusBadRequest, err)
	}
}

func updateItem(request *restful.Request, response *restful.Response) {
	itemID := request.PathParameter("book-id")
	existingItem, found := books[itemID]
	if !found {
		response.WriteError(http.StatusNotFound, restful.NewError(http.StatusNotFound, "Book not found"))
		return
	}

	updatedItem := new(Book)
	err := request.ReadEntity(updatedItem)
	if err == nil {
		// 更新 book
		existingItem.Name = updatedItem.Name
		books[itemID] = existingItem
		response.WriteEntity(existingItem)
	} else {
		response.WriteError(http.StatusBadRequest, err)
	}
}

func deleteItem(request *restful.Request, response *restful.Response) {
	itemID := request.PathParameter("book-id")
	// 删除指定 ID 的 book
	delete(books, itemID)
	response.WriteHeader(http.StatusNoContent)
}

func generateID() string {
	// 使用 UUID 生成唯一的 ID
	id := uuid.New().String()
	return "id" + id
}
