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
    googleCredsPath="./internal/adminSdkSecret.json"

    # These are the replace paths. When adding new functions, make sure to appedn to this list
    spotifyAuthReplace="github.com/pixelogicdev/gruveebackend/cmd/spotifyauth=../cmd/spotifyauth"
    tokengenReplace="github.com/pixelogicdev/gruveebackend/cmd/tokengen=../cmd/tokengen"
    socialPlatformReplace="github.com/pixelogicdev/gruveebackend/cmd/socialplatform=../cmd/socialplatform"
    createUserReplace="github.com/pixelogicdev/gruveebackend/cmd/createuser=../cmd/createuser"
    socialTokenRefreshReplace="github.com/pixelogicdev/gruveebackend/cmd/socialtokenrefresh=../cmd/socialtokenrefresh"
    createSocialPlaylistReplace="github.com/pixelogicdev/gruveebackend/cmd/createsocialplaylist=../cmd/createsocialplaylist"
    firebaseReplace="github.com/pixelogicdev/gruveebackend/pkg/firebase=../../pkg/firebase"
    socialReplace="github.com/pixelogicdev/gruveebackend/pkg/social=../../pkg/social"

    # add googleCreds to terminal instance
    export GOOGLE_APPLICATION_CREDENTIALS=$googleCredsPath

    # Go into each child directory and add replace to all mod files
    cd cmd
    for d in */
    do
        echo "Adding replace to $d"
        cd $d

        # Don't add createuser
        if [ "$d" = "createuser/" ]
        then
            go mod edit -replace $spotifyAuthReplace
            go mod edit -replace $tokengenReplace
            go mod edit -replace $socialPlatformReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $socialTokenRefreshReplace
            go mod edit -replace $createSocialPlaylistReplace
            go mod edit -replace $socialReplace
        fi

        if [ "$d" = "socialplatform/" ]
        then
            go mod edit -replace $spotifyAuthReplace
            go mod edit -replace $tokengenReplace
            go mod edit -replace $createUserReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $socialTokenRefreshReplace
            go mod edit -replace $createSocialPlaylistReplace
            go mod edit -replace $socialReplace
        fi

        if [ "$d" = "spotifyauth/" ]
        then
            go mod edit -replace $socialPlatformReplace
            go mod edit -replace $tokengenReplace
            go mod edit -replace $createUserReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $socialTokenRefreshReplace
            go mod edit -replace $createSocialPlaylistReplace
            go mod edit -replace $socialReplace
        fi
    
        if [ "$d" = "tokengen/" ] 
        then
            go mod edit -replace $socialPlatformReplace
            go mod edit -replace $spotifyAuthReplace 
            go mod edit -replace $createUserReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $socialTokenRefreshReplace
            go mod edit -replace $createSocialPlaylistReplace
            go mod edit -replace $socialReplace
        fi

        if [ "$d" = "socialtokenrefresh/" ] 
        then
            go mod edit -replace $socialPlatformReplace
            go mod edit -replace $spotifyAuthReplace 
            go mod edit -replace $createUserReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $tokengenReplace
            go mod edit -replace $createSocialPlaylistReplace
            go mod edit -replace $socialReplace
        fi

        if [ "$d" = "createsocialplaylist/" ] 
        then
            go mod edit -replace $socialPlatformReplace
            go mod edit -replace $spotifyAuthReplace 
            go mod edit -replace $createUserReplace
            go mod edit -replace $firebaseReplace
            go mod edit -replace $tokengenReplace
            go mod edit -replace $socialTokenRefreshReplace
            go mod edit -replace $socialReplace
        fi
    
        # Move back up a directory
        cd ..
    done

    # Head back to main directory and run main
    cd ..
    go run main.go
else
    echo "Make sure you are in the root of gruveebackend before running!"
fi
