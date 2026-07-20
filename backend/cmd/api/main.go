package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

var tasks = []Task{
	{
		ID:        1,
		Title:     "Learn Gin",
		Completed: false,
	},
}

func main() {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.GET("/tasks", getTasks)
		api.GET("/tasks/:id", getTaskByID)
		api.POST("/tasks", createTask)
		api.PUT("/tasks/:id", updateTask)
		api.DELETE("/tasks/:id", deleteTask)
	}

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}

func getTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
	})
}

func getTaskByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task id",
		})
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"data": task,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "task not found",
	})
}

func createTask(c *gin.Context) {
	var input Task

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	input.ID = nextTaskID()
	tasks = append(tasks, input)

	c.JSON(http.StatusCreated, gin.H{
		"data": input,
	})
}

func updateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task id",
		})
		return
	}

	var input Task

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for index := range tasks {
		if tasks[index].ID == id {
			input.ID = id
			tasks[index] = input

			c.JSON(http.StatusOK, gin.H{
				"data": tasks[index],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "task not found",
	})
}

func deleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task id",
		})
		return
	}

	for index, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:index], tasks[index+1:]...)

			c.Status(http.StatusNoContent)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "task not found",
	})
}

func nextTaskID() int {
	maxID := 0

	for _, task := range tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}

	return maxID + 1
}
