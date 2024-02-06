# ShopifyProductArchiver CLI
Tool that makes archiving large amounts of products easier, according to requirements given by the user.

# What/How
This program utilizes the product/inventory import/export scheme.

At the moment the program is a cli, which takes a product export csv and an inventory export csv as its input and creates a new product import csv according to inputs given in the program.

This program was created as a quick way to solve a specific problem and isn't expected to develop into anything bigger yet.

# Use
## Build the program with:
`go build`
## Usage:
`./main [productExport.csv] [inventoryExport.csv] [outputFile.csv]`

# Roadmap
- A GUI to be able to be used by employees, who don't have to know the codebase.
- More advanced requirements for archival
- Feature to adjust inventories
