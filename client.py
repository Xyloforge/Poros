import socket
import threading
import sys
import datetime

SERVER_IP = "127.0.0.1"
SERVER_PORT = 8080

def log(msg):
    now = datetime.datetime.now().strftime("%Y/%m/%d %H:%M:%S")
    print(f"\n[CLIENT]: {now} {msg}")

def listen_for_broadcasts(sock):
    """Background thread to continuously listen for server responses and broadcasts."""
    while True:
        try:
            data, addr = sock.recvfrom(1024)
            if not data:
                break
            log(f"Server response: {data.decode(errors='replace')}")
            # Re-print prompt for clean UI
            sys.stdout.write("Message to send: ")
            sys.stdout.flush()
        except OSError:
            # Socket closed
            break
        except Exception as e:
            log(f"Error receiving: {e}")
            break

def main():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    server_addr = (SERVER_IP, SERVER_PORT)

    # Start background listener thread for broadcasts/responses
    listener_thread = threading.Thread(target=listen_for_broadcasts, args=(sock,), daemon=True)
    listener_thread.start()

    log(f"Connected to UDP Server at {SERVER_IP}:{SERVER_PORT}")
    print("Commands:")
    print("  0         -> Create Room")
    print("  1<roomKey> -> Join Room (e.g. 1aBcDeF)")
    print("  2         -> Leave Room")
    print("  3<msg>    -> Broadcast message to room (e.g. 3Hello)")
    print("-" * 50)

    # Go serviceType enum (iota):
    # Create = 0, Join = 1, Leave = 2, Broad = 3, Unknown = 99
    service_byte_map = {
        '0': 0, # Create
        '1': 1, # Join
        '2': 2, # Leave
        '3': 3, # Broad
    }

    try:
        while True:
            msg = input("Message to send: ")
            if not msg:
                continue

            first_char = msg[0]
            if first_char in service_byte_map:
                service_byte = bytes([service_byte_map[first_char]])
                payload = service_byte + msg[1:].encode('utf-8')
            else:
                service_byte = bytes([99]) # Unknown (99)
                payload = service_byte + msg.encode('utf-8')

            sock.sendto(payload, server_addr)

    except (KeyboardInterrupt, EOFError):
        print("\nExiting UDP Client...")
    finally:
        sock.close()

if __name__ == "__main__":
    main()
