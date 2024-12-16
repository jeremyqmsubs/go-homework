package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Struct Product
type Product struct {
	ID    int
	Name  string
	Price float64
}

// Method untuk mencetak informasi produk
func (p Product) DisplayProduct() {
	fmt.Printf("Product ID: %d, Name: %s, Price: %.2f\n", p.ID, p.Name, p.Price)
}

// Method untuk mengurangi harga produk berdasarkan diskon (%)
func (p *Product) ApplyDiscount(discount float64) {
	p.Price -= p.Price * (discount / 100)
}

// Fungsi untuk menyisipkan data produk ke database
func InsertProduct(db *sql.DB, product Product, ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	query := "INSERT INTO product (name, price) VALUES (?, ?)"
	_, err := db.Exec(query, product.Name, product.Price)
	if err != nil {
		log.Printf("Failed to insert product %s: %v\n", product.Name, err)
		return
	}
	fmt.Printf("Inserted product: %s\n", product.Name)
	ch <- 1
}

// Fungsi untuk mengambil data dari database
func FetchProducts(db *sql.DB) {
	query := "SELECT id, name, price FROM product"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to fetch products: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Products in database:")
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			log.Printf("Failed to scan row: %v\n", err)
			continue
		}
		product.DisplayProduct()
	}
}

func main() {
	// Koneksi ke database
	const dsn = "root:@tcp(127.0.0.1:3306)/golang_db" // Sesuaikan dengan kredensial Anda
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer db.Close()

	// Data produk (array of Product)
	products := []Product{
		{Name: "Product A", Price: 100.00},
		{Name: "Product B", Price: 200.00},
		{Name: "Product C", Price: 300.00},
		{Name: "Product D", Price: 400.00},
		{Name: "Product E", Price: 500.00},
	}

	// Apply discount untuk setiap produk
	for i := range products {
		products[i].ApplyDiscount(10) // Terapkan diskon 10%
	}

	ch := make(chan int)
	var wg sync.WaitGroup

	go func() {
		count := 0
		for range ch {
			count++
			if count == 2 {
				FetchProducts(db)
			}
		}
	}()

	for _, product := range products {
		wg.Add(1)
		go InsertProduct(db, product, ch, &wg)
	}

	wg.Wait()
	close(ch)
	time.Sleep(1 * time.Second)
	FetchProducts(db)
}
