import json
import asyncio
from collections import defaultdict
from typing import Dict, List, Optional
import networkx as nx
from redis.asyncio import Redis
import logging

log = logging.getLogger(__name__)


class GraphManager:
    """Manages the dynamic network graph from Redis topology updates"""

    def __init__(self, redis_client: Redis):
        self.redis = redis_client
        self.graph = nx.Graph()
        self.lock = asyncio.Lock()
        self.last_update = None

    async def subscribe_to_topology(self):
        """Subscribe to topology updates from Redis"""
        pubsub = self.redis.pubsub()
        await pubsub.subscribe("topology-updates")

        async for message in pubsub.listen():
            if message["type"] == "message":
                try:
                    data = json.loads(message["data"])
                    await self.update_graph(data)
                except json.JSONDecodeError as e:
                    log.error(f"Failed to parse topology update: {e}")

    async def update_graph(self, adjacency_data: dict):
        """Update the graph with new adjacency information"""
        async with self.lock:
            self.graph.clear()

            # Build nodes and edges from adjacency
            for sat_id, entry in adjacency_data.items():
                sat_id = int(sat_id)
                self.graph.add_node(sat_id, position=tuple(entry["position"]), velocity=tuple(entry["velocity"]))

                # Add edges with distance-based weights
                for neighbor_id in entry["adjacent_satellites"]:
                    dist = self._compute_distance(entry["position"], adjacency_data[str(neighbor_id)]["position"])
                    # Weight is latency (distance / speed of light + overhead)
                    latency = dist / 3e8 + 0.001  # 1ms overhead
                    self.graph.add_edge(sat_id, neighbor_id, weight=latency, distance=dist)

            self.last_update = asyncio.get_event_loop().time()
            log.info(f"Graph updated: {self.graph.number_of_nodes()} nodes, {self.graph.number_of_edges()} edges")

    def _compute_distance(self, pos1: List[float], pos2: List[float]) -> float:
        """Euclidean distance between two positions"""
        dx = pos2[0] - pos1[0]
        dy = pos2[1] - pos1[1]
        dz = pos2[2] - pos1[2]
        return (dx**2 + dy**2 + dz**2) ** 0.5

    async def get_shortest_path(self, source: int, target: int, current_time: Optional[float] = None) -> Optional[List[int]]:
        """Find shortest path from source to target"""
        async with self.lock:
            try:
                if source not in self.graph or target not in self.graph:
                    return None

                path = nx.shortest_path(self.graph, source, target, weight="weight")
                return path
            except nx.NetworkXNoPath:
                return None
            except nx.NodeNotFound:
                return None

    async def get_path_latency(self, path: List[int]) -> float:
        """Calculate total latency for a path"""
        async with self.lock:
            total_latency = 0.0
            for i in range(len(path) - 1):
                edge_data = self.graph.get_edge_data(path[i], path[i + 1])
                if edge_data:
                    total_latency += edge_data["weight"]
            return total_latency

    async def get_hop_count(self, path: List[int]) -> int:
        """Get number of hops in a path"""
        return len(path) - 1 if path else 0

    async def is_connected(self) -> bool:
        """Check if the graph is connected"""
        async with self.lock:
            return nx.is_connected(self.graph) if self.graph.number_of_nodes() > 0 else False

    def get_network_stats(self) -> dict:
        """Get current network statistics"""
        return {
            "nodes": self.graph.number_of_nodes(),
            "edges": self.graph.number_of_edges(),
            "is_connected": nx.is_connected(self.graph) if self.graph.number_of_nodes() > 0 else False,
            "average_degree": sum(dict(self.graph.degree()).values()) / self.graph.number_of_nodes()
            if self.graph.number_of_nodes() > 0
            else 0,
        }
