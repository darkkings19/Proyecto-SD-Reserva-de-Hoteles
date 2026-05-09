from fastapi import FastAPI
from dotenv import load_dotenv

# [INICIO BLOQUE TEMPORAL] ============================================
# TODO: Eliminar la siguiente línea cuando se integre el servicio real de Reservas
from routers import testing
# [FIN BLOQUE TEMPORAL] ===============================================

# Aquí deberán importar los routers reales a futuro:
# from routers import reservations

# Load environment variables
load_dotenv()

app = FastAPI(title="Origen X - API Gateway", version="1.0.0")

# Montar routers
# [INICIO BLOQUE TEMPORAL] ============================================
# TODO: Eliminar la siguiente línea cuando se integre el servicio real de Reservas
app.include_router(testing.router, tags=["Testing (Temporal)"])
# [FIN BLOQUE TEMPORAL] ===============================================

# Aquí deberán montar los routers reales a futuro:
# app.include_router(reservations.router, prefix="/api/v1/reservations", tags=["Reservations"])

@app.get("/health")
async def health_check():
    return {"status": "ok"}
