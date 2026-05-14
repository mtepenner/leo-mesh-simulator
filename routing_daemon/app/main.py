"""Main FastAPI application for routing daemon"""

import asyncio
import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from redis.asyncio import Redis

from app.core.graph_manager import GraphManager
from app.pathfinding.connection_manager import ConnectionManager
from app.api.routes import traceroute

# Configure logging
logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)

# Global state
graph_manager = None
connection_manager = None
redis_client = None
topology_subscription_task = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Manage application startup and shutdown"""
    global graph_manager, connection_manager, redis_client, topology_subscription_task

    log.info("Starting routing daemon...")

    # Initialize Redis
    redis_addr = os.getenv("REDIS_ADDR", "redis:6379")
    redis_client = Redis.from_url(f"redis://{redis_addr}", decode_responses=True, auto_close_connection_pool=False)

    try:
        await redis_client.ping()
        log.info("✓ Connected to Redis")
    except Exception as e:
        log.error(f"Failed to connect to Redis: {e}")
        raise

    # Initialize managers
    graph_manager = GraphManager(redis_client)
    connection_manager = ConnectionManager(graph_manager)
    traceroute.set_managers(graph_manager, connection_manager)

    # Subscribe to topology updates
    topology_subscription_task = asyncio.create_task(graph_manager.subscribe_to_topology())

    # Cleanup task
    cleanup_task = asyncio.create_task(_cleanup_loop())

    log.info("✓ Routing daemon initialized")

    yield  # Application runs here

    # Shutdown
    log.info("Shutting down routing daemon...")
    topology_subscription_task.cancel()
    cleanup_task.cancel()
    await redis_client.close()
    log.info("✓ Shutdown complete")


async def _cleanup_loop():
    """Periodically clean up expired flows"""
    while True:
        try:
            await asyncio.sleep(3600)  # Every hour
            await connection_manager.cleanup_expired_flows()
        except asyncio.CancelledError:
            break
        except Exception as e:
            log.error(f"Error in cleanup loop: {e}")


# Create FastAPI app
app = FastAPI(title="LEO Mesh Routing Daemon", version="1.0.0", lifespan=lifespan)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(traceroute.router)


@app.websocket("/ws/topology")
async def websocket_topology(websocket: WebSocket):
    """WebSocket endpoint for real-time topology updates"""
    await websocket.accept()
    log.info("WebSocket client connected for topology")

    try:
        while True:
            # Send network stats periodically
            if graph_manager:
                stats = graph_manager.get_network_stats()
                await websocket.send_json(
                    {
                        "type": "stats",
                        "data": stats,
                    }
                )

            # Wait before next update
            await asyncio.sleep(1)

    except WebSocketDisconnect:
        log.info("WebSocket client disconnected")
    except Exception as e:
        log.error(f"WebSocket error: {e}")
        await websocket.close()


@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "LEO Mesh Routing Daemon",
        "version": "1.0.0",
        "endpoints": [
            "/api/traceroute/find-path",
            "/api/traceroute/path/{source}/{target}",
            "/api/traceroute/stats",
            "/api/traceroute/health",
            "/ws/topology",
        ],
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
