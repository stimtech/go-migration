# Stim Go Migration lib#

Library for database sql migrations. 

### What is this repository for? ###

Reads sql files from a folder (default db/migrations), and applies them in alphabetical order. 
Any file is only applied once. If the checksum of a file has changed since it was applied, the migration will fail.  

