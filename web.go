package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func StartHTTPServer(addr string, mm MapManager) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		state := mm.GetState()
		json.NewEncoder(w).Encode(state)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(dashboardHTML))
	})

	log.Printf("[HTTP DASHBOARD]: Listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("[HTTP DASHBOARD ERROR]: %v\n", err)
	}
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Poros Server — Matchmaking & NAT Monitor</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Google+Sans:wght@400;500;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons+Round" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0f1117;
            --surface-color: #1a1d24;
            --surface-hover: #222630;
            --card-border: #2a2e39;
            --text-primary: #f1f3f4;
            --text-secondary: #9aa0a6;
            --google-blue: #1a73e8;
            --google-blue-light: #8ab4f8;
            --google-green: #34a853;
            --google-green-light: #81c995;
            --google-yellow: #fbbc04;
            --google-red: #ea4335;
            --accent-glow: rgba(26, 115, 232, 0.15);
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Google Sans', -apple-system, BlinkMacSystemFont, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        header {
            background-color: var(--surface-color);
            border-bottom: 1px solid var(--card-border);
            padding: 16px 32px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            position: sticky;
            top: 0;
            z-index: 100;
            backdrop-filter: blur(10px);
        }

        .logo-container {
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .logo-icon {
            width: 40px;
            height: 40px;
            background: linear-gradient(135deg, var(--google-blue), #4285f4);
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #fff;
            box-shadow: 0 4px 12px var(--accent-glow);
        }

        .logo-text {
            font-size: 20px;
            font-weight: 700;
            letter-spacing: -0.5px;
            background: linear-gradient(90deg, #ffffff, #9aa0a6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .status-badge {
            display: flex;
            align-items: center;
            gap: 8px;
            background-color: rgba(52, 168, 83, 0.12);
            color: var(--google-green-light);
            padding: 6px 14px;
            border-radius: 20px;
            font-size: 13px;
            font-weight: 500;
            border: 1px solid rgba(52, 168, 83, 0.3);
        }

        .pulse-dot {
            width: 8px;
            height: 8px;
            background-color: var(--google-green);
            border-radius: 50%;
            box-shadow: 0 0 0 0 rgba(52, 168, 83, 0.7);
            animation: pulse 1.8s infinite;
        }

        @keyframes pulse {
            0% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(52, 168, 83, 0.7);
            }
            70% {
                transform: scale(1);
                box-shadow: 0 0 0 8px rgba(52, 168, 83, 0);
            }
            100% {
                transform: scale(0.95);
                box-shadow: 0 0 0 0 rgba(52, 168, 83, 0);
            }
        }

        main {
            flex: 1;
            padding: 32px;
            max-width: 1300px;
            width: 100%;
            margin: 0 auto;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 20px;
            margin-bottom: 32px;
        }

        .metric-card {
            background-color: var(--surface-color);
            border: 1px solid var(--card-border);
            border-radius: 16px;
            padding: 24px;
            display: flex;
            align-items: center;
            gap: 20px;
            transition: transform 0.2s ease, border-color 0.2s ease;
        }

        .metric-card:hover {
            transform: translateY(-2px);
            border-color: rgba(138, 180, 248, 0.4);
        }

        .metric-icon {
            width: 52px;
            height: 52px;
            border-radius: 14px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 26px;
        }

        .metric-icon.blue {
            background: rgba(26, 115, 232, 0.15);
            color: var(--google-blue-light);
        }

        .metric-icon.green {
            background: rgba(52, 168, 83, 0.15);
            color: var(--google-green-light);
        }

        .metric-icon.yellow {
            background: rgba(251, 188, 4, 0.15);
            color: var(--google-yellow);
        }

        .metric-data {
            display: flex;
            flex-direction: column;
        }

        .metric-value {
            font-size: 32px;
            font-weight: 700;
            line-height: 1.1;
        }

        .metric-label {
            font-size: 13px;
            color: var(--text-secondary);
            margin-top: 4px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            font-weight: 500;
        }

        .section-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 20px;
        }

        .section-title {
            font-size: 18px;
            font-weight: 700;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .rooms-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
            gap: 24px;
        }

        .room-card {
            background-color: var(--surface-color);
            border: 1px solid var(--card-border);
            border-radius: 16px;
            padding: 20px;
            display: flex;
            flex-direction: column;
            gap: 16px;
            transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: scale(0.97); }
            to { opacity: 1; transform: scale(1); }
        }

        .room-card:hover {
            border-color: var(--google-blue);
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
        }

        .room-card-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding-bottom: 12px;
            border-bottom: 1px solid var(--card-border);
        }

        .room-key-badge {
            font-family: 'JetBrains Mono', monospace;
            font-size: 16px;
            font-weight: 600;
            background: linear-gradient(135deg, rgba(26, 115, 232, 0.2), rgba(66, 133, 244, 0.1));
            color: var(--google-blue-light);
            padding: 6px 12px;
            border-radius: 8px;
            border: 1px solid rgba(138, 180, 248, 0.3);
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .occupancy-tag {
            font-size: 13px;
            font-weight: 500;
            color: var(--text-secondary);
        }

        .occupancy-bar-container {
            height: 6px;
            background-color: var(--card-border);
            border-radius: 3px;
            overflow: hidden;
            margin-top: 4px;
        }

        .occupancy-bar {
            height: 100%;
            background: linear-gradient(90deg, var(--google-blue), var(--google-green));
            border-radius: 3px;
            transition: width 0.4s ease;
        }

        .clients-list-title {
            font-size: 12px;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 8px;
        }

        .clients-list {
            list-style: none;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }

        .client-item {
            display: flex;
            align-items: center;
            gap: 10px;
            background-color: var(--bg-color);
            padding: 8px 12px;
            border-radius: 8px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 13px;
            border: 1px solid rgba(255, 255, 255, 0.04);
        }

        .client-avatar {
            width: 24px;
            height: 24px;
            border-radius: 50%;
            background: rgba(255, 255, 255, 0.08);
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            color: var(--google-blue-light);
        }

        .empty-state {
            grid-column: 1 / -1;
            background-color: var(--surface-color);
            border: 1px dashed var(--card-border);
            border-radius: 16px;
            padding: 48px 24px;
            text-align: center;
            color: var(--text-secondary);
        }

        .empty-icon {
            font-size: 48px;
            color: var(--card-border);
            margin-bottom: 12px;
        }

        .empty-title {
            font-size: 16px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 4px;
        }

        footer {
            text-align: center;
            padding: 24px;
            font-size: 12px;
            color: var(--text-secondary);
            border-top: 1px solid var(--card-border);
            margin-top: auto;
        }
    </style>
</head>
<body>

    <header>
        <div class="logo-container">
            <div class="logo-icon">
                <span class="material-icons-round">router</span>
            </div>
            <div>
                <div class="logo-text">Poros Server</div>
                <div style="font-size: 11px; color: var(--text-secondary);">UDP Matchmaking & NAT Traversal</div>
            </div>
        </div>
        <div class="status-badge">
            <div class="pulse-dot"></div>
            <span>LIVE MONITORING</span>
        </div>
    </header>

    <main>
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-icon blue">
                    <span class="material-icons-round">meeting_room</span>
                </div>
                <div class="metric-data">
                    <div class="metric-value" id="total-rooms">0</div>
                    <div class="metric-label">Active Rooms</div>
                </div>
            </div>

            <div class="metric-card">
                <div class="metric-icon green">
                    <span class="material-icons-round">groups</span>
                </div>
                <div class="metric-data">
                    <div class="metric-value" id="total-clients">0</div>
                    <div class="metric-label">Connected Clients</div>
                </div>
            </div>

            <div class="metric-card">
                <div class="metric-icon yellow">
                    <span class="material-icons-round">swap_calls</span>
                </div>
                <div class="metric-data">
                    <div class="metric-value">UDP :8080</div>
                    <div class="metric-label">Server Transport</div>
                </div>
            </div>
        </div>

        <div class="section-header">
            <div class="section-title">
                <span class="material-icons-round" style="color: var(--google-blue-light);">grid_view</span>
                <span>Live Room Directory</span>
            </div>
        </div>

        <div class="rooms-grid" id="rooms-container">
            <!-- Dynamic Room Cards will be injected here -->
        </div>
    </main>

    <footer>
        Poros NAT Traversal & Matchmaking Engine — Built in Go
    </footer>

    <script>
        let lastStateJSON = '';

        async function fetchState() {
            try {
                const response = await fetch('/api/state');
                const state = await response.json();
                updateDashboard(state);
            } catch (err) {
                console.error("Error fetching state:", err);
            }
        }

        function updateDashboard(state) {
            const currentStateJSON = JSON.stringify(state);
            if (currentStateJSON === lastStateJSON) {
                return; // State hasn't changed; avoid re-rendering DOM
            }
            lastStateJSON = currentStateJSON;

            document.getElementById('total-rooms').textContent = state.totalRooms || 0;
            document.getElementById('total-clients').textContent = state.totalClients || 0;

            const container = document.getElementById('rooms-container');
            const rooms = state.rooms || [];

            if (rooms.length === 0) {
                container.innerHTML = '<div class="empty-state">' +
                    '<span class="material-icons-round empty-icon">sensor_door</span>' +
                    '<div class="empty-title">No Active Rooms</div>' +
                    '<div>Create a room via client using opcode 0</div>' +
                    '</div>';
                return;
            }

            let html = '';
            for (let i = 0; i < rooms.length; i++) {
                const room = rooms[i];
                const clientCount = room.clients ? room.clients.length : 0;
                const maxClient = room.maxClient || 4;
                const percentage = Math.min((clientCount / maxClient) * 100, 100);

                let clientsHTML = '';
                if (clientCount > 0) {
                    for (let j = 0; j < room.clients.length; j++) {
                        const c = room.clients[j];
                        clientsHTML += '<li class="client-item">' +
                            '<div class="client-avatar"><span class="material-icons-round" style="font-size: 16px;">computer</span></div>' +
                            '<span>' + c + '</span>' +
                            '</li>';
                    }
                } else {
                    clientsHTML = '<div style="font-size:13px; color:var(--text-secondary);">No clients in room</div>';
                }

                html += '<div class="room-card">' +
                    '<div class="room-card-header">' +
                        '<div class="room-key-badge">' +
                            '<span class="material-icons-round" style="font-size: 18px;">vpn_key</span>' +
                            '<span>' + room.roomKey + '</span>' +
                        '</div>' +
                        '<div class="occupancy-tag">' + clientCount + ' / ' + maxClient + ' Clients</div>' +
                    '</div>' +
                    '<div class="occupancy-bar-container">' +
                        '<div class="occupancy-bar" style="width: ' + percentage + '%"></div>' +
                    '</div>' +
                    '<div>' +
                        '<div class="clients-list-title">Connected Peers (' + clientCount + ')</div>' +
                        '<ul class="clients-list">' + clientsHTML + '</ul>' +
                    '</div>' +
                '</div>';
            }
            container.innerHTML = html;
        }

        // Auto-refresh every 1 second
        setInterval(fetchState, 1000);
        fetchState();
    </script>
</body>
</html>`
