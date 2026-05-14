"""Dynamic Dijkstra-based pathfinding for mesh routing"""

import asyncio
from typing import List, Optional
import networkx as nx
import logging

log = logging.getLogger(__name__)


class DynamicDijkstra:
    """Dijkstra pathfinding optimized for dynamic networks"""

    @staticmethod
    async def find_path(graph: nx.Graph, source: int, target: int) -> Optional[List[int]]:
        """
        Find shortest path using Dijkstra algorithm
        Runs in thread pool to avoid blocking
        """
        loop = asyncio.get_event_loop()
        try:
            path = await loop.run_in_executor(None, lambda: nx.shortest_path(graph, source, target, weight="weight"))
            return path
        except (nx.NetworkXNoPath, nx.NodeNotFound) as e:
            log.debug(f"No path from {source} to {target}: {e}")
            return None

    @staticmethod
    def find_path_sync(graph: nx.Graph, source: int, target: int) -> Optional[List[int]]:
        """Synchronous version for use in thread pool"""
        try:
            return nx.shortest_path(graph, source, target, weight="weight")
        except (nx.NetworkXNoPath, nx.NodeNotFound):
            return None


class ConnectionManager:
    """Manages active connections and rerouting logic"""

    def __init__(self, graph_manager):
        self.graph_manager = graph_manager
        self.active_flows = {}  # (source, target) -> path history
        self.reroute_threshold = 1.5  # Trigger reroute if latency > 1.5x optimal
        self.lock = asyncio.Lock()

    async def establish_route(self, source: int, target: int) -> Optional[dict]:
        """Establish a route from source to target"""
        path = await self.graph_manager.get_shortest_path(source, target)

        if not path:
            return None

        latency = await self.graph_manager.get_path_latency(path)
        hop_count = await self.graph_manager.get_hop_count(path)

        flow_key = (source, target)
        async with self.lock:
            self.active_flows[flow_key] = {
                "path": path,
                "latency": latency,
                "hop_count": hop_count,
                "reroute_count": 0,
                "last_update": asyncio.get_event_loop().time(),
            }

        return {
            "source": source,
            "target": target,
            "path": path,
            "latency_ms": latency * 1000,
            "hop_count": hop_count,
        }

    async def check_and_reroute(self, source: int, target: int) -> Optional[dict]:
        """Check if rerouting is needed and perform if necessary"""
        flow_key = (source, target)

        async with self.lock:
            if flow_key not in self.active_flows:
                return None

            flow_info = self.active_flows[flow_key]
            current_path = flow_info["path"]
            old_latency = flow_info["latency"]

        # Check if current path still exists
        new_path = await self.graph_manager.get_shortest_path(source, target)

        if new_path is None:
            # No path available
            async with self.lock:
                if flow_key in self.active_flows:
                    del self.active_flows[flow_key]
            return None

        new_latency = await self.graph_manager.get_path_latency(new_path)

        # Check if we need to reroute
        if new_latency < old_latency / self.reroute_threshold:
            # Found significantly better path
            hop_count = await self.graph_manager.get_hop_count(new_path)

            async with self.lock:
                if flow_key in self.active_flows:
                    self.active_flows[flow_key].update(
                        {
                            "path": new_path,
                            "latency": new_latency,
                            "hop_count": hop_count,
                            "reroute_count": flow_info["reroute_count"] + 1,
                            "last_update": asyncio.get_event_loop().time(),
                        }
                    )

            log.info(f"Rerouted {source}->{target}: {len(current_path)-1} hops -> {len(new_path)-1} hops")

            return {
                "source": source,
                "target": target,
                "path": new_path,
                "latency_ms": new_latency * 1000,
                "hop_count": hop_count,
                "reroute_count": flow_info["reroute_count"] + 1,
            }

        return None

    async def get_active_route(self, source: int, target: int) -> Optional[dict]:
        """Get current route for a flow"""
        flow_key = (source, target)
        async with self.lock:
            if flow_key not in self.active_flows:
                return None

            flow = self.active_flows[flow_key]
            return {
                "source": source,
                "target": target,
                "path": flow["path"],
                "latency_ms": flow["latency"] * 1000,
                "hop_count": flow["hop_count"],
                "reroute_count": flow["reroute_count"],
            }

    async def cleanup_expired_flows(self, timeout: float = 3600):
        """Remove flows that haven't been updated recently"""
        current_time = asyncio.get_event_loop().time()
        async with self.lock:
            expired = [
                flow_key
                for flow_key, flow in self.active_flows.items()
                if current_time - flow["last_update"] > timeout
            ]

            for flow_key in expired:
                del self.active_flows[flow_key]
                log.info(f"Cleaned up expired flow: {flow_key}")
