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

Next, you'll need to setup and run this (Github) project by:

1. Place the downloaded credentials using name:f `firebase_credentials.json`  on `backend/config`
