package main

import (
	"os"
	"tp1/common/middleware"
)

func main() {
	producer := middleware.NewProducer("trip_counter", 1, false)
	consumer := middleware.NewConsumer("year_filter", "", 1, os.Getenv("ID"))

	yearFilter := NewYearFilter(producer, consumer)
	yearFilter.Run()
}
