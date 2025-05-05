#!/bin/bash

# First apply the CRD
echo "Applying Game CRD..."
kubectl apply -f 03-game-crd.yaml

# Wait for CRD to be established
echo "Waiting for CRD to be established..."
sleep 5

# Try to apply each invalid example
echo -e "\nTrying invalid examples..."
echo "----------------------------------------"

# Split the examples file into separate files
csplit -f "invalid-" -b "%02d.yaml" 05-game-invalid-examples.yaml '/^---$/' '{*}'

# Try each invalid example
for file in invalid-*; do
    echo -e "\nTrying to apply $file:"
    kubectl apply -f "$file"
    echo "----------------------------------------"
done

# Cleanup
rm invalid-*