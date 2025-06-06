# Production Readiness

## How would you scale this service to handle 10,000+ concurrent users?

- Multiple instances behind a load balancer;
- Set up auto-scale instances based on CPU, memory and number of connections;
- One leader instance to fetch prices and dispatch internal events;
- Internal events are processed by every instance and save in-memory data to have fast access to latest data;
- Use Redis or MongoDB to persist long-lived price updates. If an instance does not have all data in-memory, it can fetch older data from database. Both Redis (sorted set) and MongoDB provide good support for timeseries storage.

## How would you ensure reliability, fault-tolerance, and observability?

- Collect key metrics (cpu, memory, network) using an APM tool;
- Collect the number of connections using Prometheus;
- Use Grafana to visualize all metrics, and structured logs, in one place. Create alerts and dashboards for quick visualization;
- Distributed traces to easily find out possible issues;
- Use OpenTelemetry with Grafana would cover all these needs.
