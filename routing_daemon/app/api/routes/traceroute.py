from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel
from typing import List, Optional
import logging

log = logging.getLogger(__name__)

router = APIRouter(prefix="/api/traceroute", tags=["traceroute"])

# Global reference (will be injected via dependency)
_graph_manager = None
_connection_manager = None


def set_managers(graph_manager, connection_manager):
    global _graph_manager, _connection_manager
    _graph_manager = graph_manager
    _connection_manager = connection_manager


class TraceRequest(BaseModel):
    source: int
    target: int


class TraceResponse(BaseModel):
    source: int
    target: int
    path: List[int]
    latency_ms: float
    hop_count: int
    reroute_count: int = 0


class NetworkStats(BaseModel):
    nodes: int
    edges: int
    is_connected: bool
    average_degree: float


@router.post("/find-path", response_model=TraceResponse)
async def find_path(request: TraceRequest):
    """Find the shortest path from source to target satellite"""
    if not _graph_manager:
        raise HTTPException(status_code=503, detail="Graph not yet initialized")

    # Establish or update route
    route = await _connection_manager.establish_route(request.source, request.target)

    if not route:
        raise HTTPException(status_code=404, detail=f"No path from {request.source} to {request.target}")

    return TraceResponse(**route)


@router.get("/path/{source}/{target}", response_model=TraceResponse)
async def get_path(source: int, target: int):
    """Get current route between two satellites"""
    if not _graph_manager:
        raise HTTPException(status_code=503, detail="Graph not yet initialized")

    route = await _connection_manager.get_active_route(source, target)

    if not route:
        # Try to establish a new route
        route = await _connection_manager.establish_route(source, target)

    if not route:
        raise HTTPException(status_code=404, detail=f"No path from {source} to {target}")

    return TraceResponse(**route)


@router.post("/reroute/{source}/{target}")
async def trigger_reroute(source: int, target: int):
    """Trigger re-evaluation of a route"""
    if not _graph_manager:
        raise HTTPException(status_code=503, detail="Graph not yet initialized")

    result = await _connection_manager.check_and_reroute(source, target)

    if result:
        return {"status": "rerouted", "route": TraceResponse(**result)}
    else:
        current = await _connection_manager.get_active_route(source, target)
        if current:
            return {"status": "no_change", "route": TraceResponse(**current)}
        else:
            raise HTTPException(status_code=404, detail=f"No path from {source} to {target}")


@router.get("/stats", response_model=NetworkStats)
async def get_network_stats():
    """Get current network statistics"""
    if not _graph_manager:
        raise HTTPException(status_code=503, detail="Graph not yet initialized")

    stats = _graph_manager.get_network_stats()
    return NetworkStats(**stats)


@router.get("/health")
async def health():
    """Health check endpoint"""
    if not _graph_manager:
        return {"status": "initializing"}

    stats = _graph_manager.get_network_stats()
    return {
        "status": "healthy",
        "nodes": stats["nodes"],
        "edges": stats["edges"],
        "connected": stats["is_connected"],
    }
