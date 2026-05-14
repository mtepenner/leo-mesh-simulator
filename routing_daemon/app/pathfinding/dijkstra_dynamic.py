"""Dijkstra algorithm module"""

import logging

log = logging.getLogger(__name__)


class DijkstraAlgorithm:
    """Custom Dijkstra implementation for educational purposes"""

    @staticmethod
    def shortest_path(graph, source, target):
        """
        Find shortest path using Dijkstra algorithm
        Graph is a dict of {node: {neighbor: weight, ...}}
        """
        if source not in graph or target not in graph:
            raise ValueError(f"Source or target not in graph")

        distances = {node: float("inf") for node in graph}
        distances[source] = 0
        previous = {node: None for node in graph}
        unvisited = set(graph.keys())

        while unvisited:
            # Find unvisited node with minimum distance
            current = min(unvisited, key=lambda node: distances[node])

            if distances[current] == float("inf"):
                break  # Unreachable nodes

            if current == target:
                # Reconstruct path
                path = []
                node = target
                while node is not None:
                    path.append(node)
                    node = previous[node]
                return list(reversed(path))

            unvisited.remove(current)

            # Update distances to neighbors
            for neighbor, weight in graph.get(current, {}).items():
                if neighbor in unvisited:
                    new_distance = distances[current] + weight
                    if new_distance < distances[neighbor]:
                        distances[neighbor] = new_distance
                        previous[neighbor] = current

        raise ValueError(f"No path from {source} to {target}")
