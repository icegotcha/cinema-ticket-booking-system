# Cinema Ticket Booking System


## System Architecture

## Tech Stack

- Backend
    - Language: Go
    - Framework: Gin
    - Database: MongoDB, Redis (For Distributed Locking)
    - Message Queue: Kafka 
- Frontend (Both Auth & Admin Side): Vue 3
- Auth: Firebase Auth
- Deployment: Docker

## Flow

[TODO]

## Setup

Before you begin, you'll need to create a firebase project [here](https://console.firebase.google.com/). In the new project, go to the menu: Settings > Service accounts > Firebase Admin  SDK, then generate and downloaded new private key

Next, you'll need to setup this (Github) project

1. Place the downloaded credentials on the root (or anywhere if you wish eg. `backend/internal/config`) of the backend project

2. Set the correct credentials file name and location including the other environment varriables in `.env` file and place it on the same location in the step 1 (example file: `backend/.env.example`)