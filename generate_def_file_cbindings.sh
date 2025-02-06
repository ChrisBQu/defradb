#!/bin/bash

# Path to the header file
HEADER_FILE="build/libdefradb.h"
DEF_FILE="build/libdefradb.def"

# Start the .def file with the LIBRARY and EXPORTS lines
echo "LIBRARY libdefradb" > $DEF_FILE
echo "EXPORTS" >> $DEF_FILE

# Extract functions marked with __declspec(dllexport)
# This will match function declarations after __declspec(dllexport) and capture the function name
grep -oP '__declspec\(dllexport\)\s+.*\s+(\w+)\(' $HEADER_FILE | \
sed -E 's/.*\s+(\w+)\(.*/\1/' >> $DEF_FILE

echo "DEF file generated at $DEF_FILE"