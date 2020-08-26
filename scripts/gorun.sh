#!/bin/bash
# This script is used for building and running the gruveebackend locally. 
# The main point of it is to append "replace" statements to all the mod files when building developing locally.
# Currently it just dumps every firebase function into each mod file, but should be updated.
# A deployment script for each firebase function should also be made that removes the replace lines.

echo "Starting Go Run Script..."

# Check if in root of project
if [[ -f "main.go" ]]
then
    # Variables
    googleCredsPath="./internal/adminSdkSecret-Dev.json"

    # These are the replace paths that are needed for shared packages.
    firebaseReplace="github.com/pixelogicdev/gruveebackend/pkg/firebase=../../pkg/firebase"
    socialReplace="github.com/pixelogicdev/gruveebackend/pkg/social=../../pkg/social"
    sawmillReplace="github.com/pixelogicdev/gruveebackend/pkg/sawmill=../../pkg/sawmill"
    mediaHelpersReplace="github.com/pixelogicdev/gruveebackend/pkg/mediahelpers=../../pkg/mediahelpers"

    # Add googleCreds to terminal instance
    export GOOGLE_APPLICATION_CREDENTIALS=$googleCredsPath

    # Go into each child directory and add replace to all mod files
    cd cmd
    for d in */
    do
        echo "Adding replace to $d"
        cd $d

        go mod edit -replace $firebaseReplace
        go mod edit -replace $socialReplace
        go mod edit -replace $sawmillReplace
        go mod edit -replace $mediaHelpersReplace

        # Move back up a directory
        cd ..
    done

    # Head back to main directory and run main
    cd ..
    go run main.go
else
    echo "Make sure you are in the root of gruveebackend before running!"
fi
