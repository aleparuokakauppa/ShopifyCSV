# ShopifyProductArchiver
An interactive CLI tool that makes archiving large amounts of products easier, according to requirements given by the user.

# Why / What
This program makes archiving large amounts of products a lot easier, since the this functionality is mainly offered in Shopify through apps.
Mass archival is useful, when the inventory of a store rotates a lot.
The program relies on the [Shopify import/export scheme](https://help.shopify.com/en/manual/products/import-export).

The program reads from a `product export CSV` and an `inventory export CSV`. From these it generates a new CSV in the desired directory. 
Running the program prompts the user to give a `target sales channel`, `tags` that should be archived, and a `minimum stock` for products to not be archived.

# Build and run
## Build the program with:
```bash
go build
```
## Run it with
```bash
./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]
```
Replace the fields in the brackets with the appropriate filepaths.

### Note
The source code is somewhat easily modifiable within the `archiveWithTags`-function to act according to customized behaviour.

# Roadmap
- A GUI to be able to be used by employees, who don't have to know the codebase.
- More advanced requirements for archival
- Feature to adjust inventories
