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
These are hardcoded because the import/export expects these indexes
*/
const (
    // Hardcoded product CSV indexes related to its header
    handleIndex = 0
    titleIndex = 1
    vendorIndex = 3
    productCategoryIndex = 4
    tagsIndex = 6
    publishedIndex = 7
    variantPriceIndex = 19
    statusIndex = 48
    // - Inventory indexes

    skuIndex = 8
    // The first sales channel is at index 11
    salesChannelSindex = 11
)

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
func (s Product) formatToShopifyProduct() []string {
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

    if len(os.Args) > 1 && os.Args[1] == "-h"{
        println("This program is intended for use with Shopify product and inventory export files.")
        println("Usage:")
        println("   ./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]")
        println()
        println("The program checks products that have a certain tag attached to them and archives them if they don't have enough inventory left")
        println("The resulting file should be imported in Shopify admin")
        println("The field 'Overwrite products with matching handles' should be enabled in shopify for changes to apply")
        return
    } 

    if len(os.Args) < 4 {
        fmt.Printf("Not enough arguments. Need 3, got %d.\n", len(os.Args)-1)
        println("Usage:")
        println("   ./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]")
        println()
        println("Give -h as the first argument for help")
        return
    }

    productExportFileName = os.Args[1]
    inventoryExportFileName = os.Args[2]
    outputFileName = os.Args[3]

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

    productCSVheader, err := productsReader.Read()
    if err != nil {
        panic(fmt.Errorf("Product CSV-header couldn't be read: %v", err))
    }
    inventoryCSVheader, err := inventoryReader.Read()
    if err != nil {
        panic(fmt.Errorf("Inventory CSV-header couldn't be read: %v", err))
    }

    salesChannel := getSalesChannel(inventoryCSVheader)

    // Get products from CSV readers
    products, err := getProducts(productsReader, inventoryReader, salesChannel)
    if err != nil {
        panic(err.Error())
    }

    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("What tags should be archived?")
    scanner.Scan()
    illegalTags := strings.Fields(scanner.Text())

    fmt.Println("Give minumum stock to not be archived (if less than this value, it gets archived)")
    scanner.Scan()
    minStock, err := strconv.Atoi(scanner.Text())
    if err != nil {
        fmt.Printf("Given minimum stock isn't an integer.\n")
        panic(err)
    }

    timeStart := time.Now()
    fmt.Printf("Starting task @ %v\n", timeStart)

    archivedProducts := archiveWithTags(products, illegalTags, minStock)

    var rawProductSlice [][]string
    for _, product := range archivedProducts {
        rawProductSlice = append(rawProductSlice, product.formatToShopifyProduct())
    }

    writeCSV(rawProductSlice, productCSVheader, outputFileName)

    timeDelta := time.Now().UnixMilli() - timeStart.UnixMilli()
    fmt.Printf("Task finished @ %v, took %d milliseconds\n", time.Now(), timeDelta)
}

/*
Returns the sales channel index in the inventoryCSVheader
*/
func getSalesChannel(inventoryCSVheader []string) (channel int) {
    var channels []string
    for i := salesChannelSindex; i < len(inventoryCSVheader)-1; i++ {
        channels = append(channels, inventoryCSVheader[i])
    }
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Println("What sales channel should be used?")
    for i, channel := range channels {
        fmt.Printf("%d: %s\n", i, channel)
    }
    // Loop for getting a proper ID
    for {
        fmt.Println("Give channel ID:")
        var err error
        scanner.Scan()
        channel, err = strconv.Atoi(scanner.Text())
        if err == nil {
            break
        }
        fmt.Println("Given channel isn't an integer. Try again")
    }
    return channel
}


/*
Writes the data as new lines into a CSV file with csvHeader as its headers.

Filename is the new CSV-file's name
*/
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

/*
Returns products that have `tags` and which have less stock than minStock
*/
func archiveWithTags(products []Product, tags []string, minStock int) (arcivedProducts []Product) {
    for _, product := range products {
        for _, tag := range tags {
            if product.hasTag(tag) && product.Inventory < minStock{
                product.Status = "archived"
                arcivedProducts = append(arcivedProducts, product)
            }
        }
    }
    return arcivedProducts
}

/*
Returns all products read from pReader with its inventory values read from iReader
*/
func getProducts(pReader *csv.Reader, iReader *csv.Reader, salesChannel int) (products []Product, err error) {
    // Append the inventory into inventoryRecord
    inventoryRecord, err := iReader.ReadAll()
    if err != nil {
        return products, fmt.Errorf("Inventory couldn't be read: %v", err)
    }
    productRecord, err := pReader.ReadAll()
    if err != nil {
        return products, fmt.Errorf("Products couldn't be read: %v", err)
    }

    for _, rawProduct := range productRecord {
        product := formatToProduct(rawProduct)
        product.Inventory, err = getInventory(product.Handle, salesChannel, inventoryRecord)
        if err != nil {
            if err == fmt.Errorf("404") {
                return products, fmt.Errorf("Product %s handle not found in the inventoryRecord", product.Handle)
            } else {
                return products, err
            }
        }
        products = append(products, product)
    }

    return products, nil
}

/*
Assigns the values from hardcoded indexes into the product
*/
func formatToProduct(rawProduct []string) (product Product) {
    product.Handle = rawProduct[handleIndex]
    product.Title = rawProduct[titleIndex]
    product.Status = rawProduct[statusIndex]
    product.Tags = strings.Split(rawProduct[tagsIndex], ", ")
    return product
}

/*
Gets the value (inventory) at salesChannel index if handle is found
*/
func getInventory(handle string, salesChannel int, inventoryRecord [][]string) (inventory int, err error) {
    // Correct for the index offset
    salesChannel += salesChannelSindex
    for _, record := range inventoryRecord {
        if handle == record[handleIndex] {
            if record[salesChannel] == "not stocked" {
                return 0, nil
            }
            inventory, err = strconv.Atoi(record[salesChannel])
            if err != nil {
                return 0, fmt.Errorf("Cannot convert inventory stock value found with handle: %s", handle);
            }
            return inventory, nil
        }
    }
    return -1, nil
}
