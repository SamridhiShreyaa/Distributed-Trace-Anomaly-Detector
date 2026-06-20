# Demo Walkthrough

## Setup

1. Build and start the stack:
   ```bash
   make demo-up
   ```

2. Verify services are running:
   - Detector API: http://localhost:8080
   - Grafana: http://localhost:3000 (admin/admin)

## Running the Demo

1. In a terminal, monitor traces:
   ```bash
   watch -n 1 'curl -s http://localhost:8080/traces | head -20'
   ```

2. In another terminal, inject a regression:
   ```bash
   make inject-regression
   ```

3. Watch the detector identify the root cause in Grafana or via curl

4. Clean up:
   ```bash
   make demo-down
   ```

## Expected Output

When a regression is detected, you should see:
- Anomaly alert in Slack
- New trace data in storage
- Updated baseline statistics
- Root-cause span highlighted in Grafana
