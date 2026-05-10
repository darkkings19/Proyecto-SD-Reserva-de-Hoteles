from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv
from routers import reservations

# Load environment variables
load_dotenv()

app = FastAPI(title="Origen X - API Gateway", version="1.0.0")

# CORS: permitir que el frontend (abierto como archivo local) se comunique
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Montar routers
app.include_router(reservations.router, tags=["Reservations"])

@app.get("/health")
async def health_check():
    return {"status": "ok"}
