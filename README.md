# HARD DEADLINE: 30TH JUNE 2023

## Hueify

This is the start of the new hue

![hue-demo](./hueify_demo.png)

#### How to Run
To build the frontend and run the server: `./run.sh`

To just run the server: `go run main.go`

#### API reference

https://app.swaggerhub.com/apis/amirrezapanahi/hueify/1.0.0#/developers/NewAlbums

#### Business Model
https://docs.google.com/document/d/1eVjes-pO3xvfwdHUF0d7mZ8O56oZ-w6KAXtYMC7PvaQ/edit?usp=sharing

#### Other Ideas
- Feed in a color palette and generate playlist from that (Updated daily)
    - Generate custom artwork for playlist
        - Give users color wheel and pick 2 dominant colors
        - Gradient ball with 2 most dominant colors
    - Generated playlist will have color palette
    - If want to add other songs:
        - Input playlist link into hueify
        - Get color palette from playlist cover
        - Remove songs already in playlist from recommendations generated
        - Add new recommendations to existing playlist
    - Additional requirements: authentication methods
- Once recommended albums generated, take 3+ songs from each album and create a playlist

Use placeholder loading for the recommended albums to notify the user that albums are being searched and prevents the ambiguity of something being broken 
