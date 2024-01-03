package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
Product csv indexes:
Handle // Index 0
Title // Index 1
Vendor // Index 3
Product_Category // Index 4
Tags // Index 6
Published // Index 7
Variant_Price // Index 19
Status // Index 48

Inventory csv indexes:
Handle // Index 0
Title // Index 1
SKU // Index 8
Sales channels begin at 11:
Leluaitta_Varastomyymälä // Index 11
Tinks_Juhlakauppa // Index 12
Wolt // Index 13
*/

type Product struct {
    Handle string // Index 0
    Title string // Index 1
    Vendor string // Index 3
    Product_Category string // Index 4
    Tags []string // Index 6
    Published string // Index 7
    Variant_Price string // Index 19
    Status string // Index 48
    Inventory int
}

// Returns the Product's properties as a string slice
// formatted with Shopify product indexes
func (s Product) formatShopifyProduct() []string {
    // Fill the slice with data at their appropriate indexes
    var result []string
    result = append(result, s.Handle) // Index 0
    result = append(result, s.Title) // Index 1
    result = append(result, "") // Index 2
    result = append(result, s.Vendor) // Index 3
    result = append(result, s.Product_Category) // Index 4
    result = append(result, "") // Index 5
    result = append(result, strings.Join(s.Tags, ", ")) // Index 6
    result = append(result, s.Published) // Index 7

    for index := 8; index <= 18; index++ {
        result = append(result, "") // Indexes through 8 to 18
    }
    result = append(result, s.Variant_Price) // Index 19

    for index := 20; index <= 47; index++ {
        result = append(result, "") // Indexes through 20 to 47
    }

    result = append(result, s.Status) // 48
    // Inventory is in a different file
    return result
}


func (product Product) hasTag(targetTag string) (result bool) {
    for _, tag := range product.Tags {
        if tag == targetTag {
            return true
        }
    }
    return false
}

func main() {
    var productExportFileName string
    var inventoryExportFileName string
    var outputFileName string

    if len(os.Args) > 1 {
        if os.Args[1] == "-h" {
            println("This program is intended for use with Shopify product and inventory export files.")
            println("Usage:")
            println("   ./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]")
            println()
            println("In this version source code must be modified for intended functionality")
            println("It checks products that have a certain tag attached to them and archives them if they don't have inventory left")
            println("The resulting file should be imported in Shopify admin")
            println("The field 'Overwrite products with matching handles' should be enabled for changes to apply")
            return
        }
    } 
    if len(os.Args) < 4 {
        fmt.Printf("Not enough arguments. Need 3, got %d.\n", len(os.Args)-1)
        println("Usage:")
        println("   ./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]")
        println()
        println("Give -h as the first argument for help")
        return
    } else {
        productExportFileName = os.Args[1]
        inventoryExportFileName = os.Args[2]
        outputFileName = os.Args[3]
    }

    timeStart := time.Now()
    fmt.Printf("Starting task @ %v\n", timeStart)
    // Open the required files
    productExportFile, err := os.Open(productExportFileName)
    if err != nil {
        panic(err.Error())
    }
    defer productExportFile.Close()

    inventoryExportFile, err := os.Open(inventoryExportFileName)
    if err != nil {
        panic(err.Error())
    }
    defer inventoryExportFile.Close()

    productsReader := csv.NewReader(productExportFile)
    inventoryReader := csv.NewReader(inventoryExportFile)

    // Read the CSV headers
    productCSVheader, err := productsReader.Read()
    if err != nil {
        fmt.Printf("Products File couldn't be read\n")
        panic(err.Error())
    }
    _, err = inventoryReader.Read()
    if err != nil {
        fmt.Printf("Inventory File couldn't be read\n")
        panic(err.Error())
    }
    // Get products from CSV readers
    products, err := getProducts(productsReader, inventoryReader)
    if err != nil {
        panic(err.Error())
    }

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Println("What tags should be archived?")
    scanner.Scan()
    illegalTagsString := scanner.Text()
    illegalTags := strings.Fields(illegalTagsString)

    fmt.Println("What should be the max stock for these products?")
    scanner.Scan()
    maxStockString := scanner.Text()
    maxStock, err := strconv.Atoi(maxStockString)
    if err != nil {
        fmt.Printf("Given max stock isn't an integer.\n")
        panic(err.Error())
    }

    archivedProducts := archiveWithTags(products, illegalTags, maxStock)

    if err != nil {
        panic(err.Error())
    }

    var productStringSlice [][]string
    for _, product := range archivedProducts {
        productStringSlice = append(productStringSlice, product.formatShopifyProduct())
    }

    writeCSV(productStringSlice, productCSVheader, outputFileName)

    timeDelta := time.Now().UnixMilli() - timeStart.UnixMilli()
    fmt.Printf("Task finished @ %v, took %d milliseconds\n", time.Now(), timeDelta)
}


// Writes the data as new lines into a CSV file with csvHeader as its headers.
// filename is the new CSV-file's name
func writeCSV(data [][]string, csvHeader []string, filename string) error {
    newCSV, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer newCSV.Close()
    csvWriter := csv.NewWriter(newCSV)
    csvWriter.Write(csvHeader)
    fmt.Printf("writing data...\n")
    csvWriter.WriteAll(data)
    return nil
}

func archiveWithTags(products []Product, tags []string, maxStock int) (arcivedProducts []Product) {
    for _, product := range products {
        for _, tag := range tags {
            if product.hasTag(tag) && product.Inventory < maxStock{
                product.Status = "archived"
                arcivedProducts = append(arcivedProducts, product)
            }
        }
    }
    return arcivedProducts
}

func getProducts(productsReader *csv.Reader, inventoryReader *csv.Reader) (products []Product, err error) {
    // Append the inventory into inventoryRecord
    inventoryRecord, err := inventoryReader.ReadAll()
    if err != nil {
        return products, err
    }
    productRecord, err := productsReader.ReadAll()
    if err != nil {
        return products, err
    }

    for _, product := range productRecord {
        formattedProduct, err := getProductFormatted(product, inventoryRecord)
        if err != nil {
            return products, err
        }
        products = append(products, formattedProduct)
    }

    return products, nil
}

func getProductFormatted(unformattedProduct []string, inventoryRecord [][]string) (product Product, err error) {
    // Append the values into the new product
    product.Handle = unformattedProduct[0]
    product.Title = unformattedProduct[1]
    product.Status = unformattedProduct[48]
    product.Tags = strings.Split(unformattedProduct[6], ", ")

    // Check the product's inventory
    productInventoryString := getInventory(product.Handle, inventoryRecord)
    if err != nil {
        if productInventoryString != "not stocked" {
            return product, err
        } else {
            product.Inventory = 0
        }
    } else {
        product.Inventory, err = strconv.Atoi(productInventoryString)
        if err != nil {
            return product, err
        }
        if product.Inventory == -404 {
            // Handle not found in the inventory file
        }
    }
    return product, nil
}

// Gets the value (inventory) at index 12 if handle is found
// If not found, returned inventory is -404 not found
func getInventory(handle string, inventoryRecord [][]string) (inventory string) {
    for _, record := range inventoryRecord {
        if handle == record[0] {
            if record[12] == "not stocked" {
                return "0"
            } else {
                return record[12]
            }
        }
    }
    return "-404"
}
