#!/bin/bash
set -e

echo "=== Testing status ==="
./bin/storage-bridge status

echo -e "\n=== Testing pwd ==="
./bin/storage-bridge pwd

echo -e "\n=== Testing mkdir ==="
./bin/storage-bridge mkdir test-folder-all

echo -e "\n=== Testing cd (into folder) ==="
./bin/storage-bridge cd test-folder-all
./bin/storage-bridge pwd

echo -e "\n=== Testing upload ==="
echo "hello from the automated test" > temp-test-file.txt
./bin/storage-bridge upload temp-test-file.txt

echo -e "\n=== Testing list ==="
./bin/storage-bridge list

echo -e "\n=== Testing cat ==="
./bin/storage-bridge cat temp-test-file.txt

echo -e "\n=== Testing download ==="
./bin/storage-bridge download temp-test-file.txt downloaded-test-file.txt
cat downloaded-test-file.txt
rm downloaded-test-file.txt

echo -e "\n=== Testing move (rename) ==="
./bin/storage-bridge move temp-test-file.txt renamed-file.txt
./bin/storage-bridge list

echo -e "\n=== Testing remove (renamed file) ==="
./bin/storage-bridge remove renamed-file.txt --yes

echo -e "\n=== Testing cd (back to root) ==="
./bin/storage-bridge cd /
./bin/storage-bridge pwd

echo -e "\n=== Testing remove (folder) ==="
./bin/storage-bridge remove test-folder-all --yes

echo -e "\n=== Testing providers list ==="
./bin/storage-bridge providers

echo -e "\n=== Cleaning up ==="
rm temp-test-file.txt
echo "All tests passed successfully!"
