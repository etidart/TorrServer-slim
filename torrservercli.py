import json
import sys
import requests
import cmd
from collections import defaultdict
from urllib.parse import urljoin

class TorrServerCLI(cmd.Cmd):
    prompt = "(torrserver) "
    
    def __init__(self, base_url):
        super().__init__()
        self.base_url = base_url
        self.current_torrent = None
        self.current_path = []
        self.tree = {}
        
    def _api_request(self, endpoint, method='post', json_data=None):
        url = urljoin(self.base_url, endpoint)
        try:
            if method.lower() == 'post':
                response = requests.post(url, json=json_data)
            else:
                response = requests.get(url, params=json_data)
                
            if response.status_code == 200:
                return response.json()
            print(f"Error {response.status_code}: {response.text}")
        except requests.exceptions.ConnectionError:
            print("Connection failed. Is Torrserver running?")
        return None

    def do_add(self, arg):
        """Add torrent by magnet link: add <magnet_link>"""
        if not arg:
            print("Please provide a magnet link")
            return
            
        data = {
            "action": "add",
            "link": arg,
            "save_to_db": True
        }
        result = self._api_request("/torrents", json_data=data)
        if result:
            print("Torrent added successfully!")

    def do_list(self, arg):
        """List available torrents"""
        data = {"action": "list"}
        torrents = self._api_request("/torrents", json_data=data)
        if torrents:
            print("\nAvailable Torrents:")
            for i, torrent in enumerate(torrents):
                print(f"{i+1}. {torrent.get('title') or torrent.get('name')}")
                print(f"   Hash: {torrent['hash']}")
                print(f"   Status: {torrent['stat_string']}")
                print(f"   Size: {torrent['torrent_size'] / (1024**2):.2f} MB\n")

    def do_select(self, arg):
        """Select a torrent: select <torrent_index>"""
        try:
            index = int(arg) - 1
            data = {"action": "list"}
            torrents = self._api_request("/torrents", json_data=data)
            if not torrents or index < 0 or index >= len(torrents):
                print("Invalid torrent index")
                return
                
            self.current_torrent = torrents[index]
            print(f"Selected torrent: {self.current_torrent['title']}")
            self._build_file_tree()
            self.current_path = []
            self.do_ls("")
        except ValueError:
            print("Please enter a valid number")

    def _build_file_tree(self):
        self.tree = defaultdict(dict)
        for file in json.loads(self.current_torrent['data'])["TorrServer"]["Files"]:
            path = file['path'].split('/')
            current = self.tree
            for part in path[:-1]:
                current = current.setdefault(part, defaultdict(dict))
            current[path[-1]] = file

    def do_ls(self, arg):
        """List files in current directory"""
        if not self.current_torrent:
            print("No torrent selected. Use 'select' first.")
            return
            
        current_dir = self.tree
        for part in self.current_path:
            current_dir = current_dir.get(part, {})
            
        print("\nCurrent directory:")
        for i, item in enumerate(sorted(current_dir.keys()), 1):
            if isinstance(current_dir[item], dict) and 'id' not in current_dir[item]:
                print(f"{i}. [DIR] {item}/")
            else:
                size = current_dir[item]['length'] / (1024**2)
                print(f"{i}. {item} ({size:.2f} MB)")

    def do_cd(self, arg):
        """Change directory: cd <dir_index> or 'cd ..' to go up"""
        if not self.current_torrent:
            print("No torrent selected")
            return
            
        current_dir = self.tree
        for part in self.current_path:
            current_dir = current_dir.get(part, {})
            
        if arg == "..":
            if self.current_path:
                self.current_path.pop()
            self.do_ls("")
            return
            
        try:
            index = int(arg) - 1
            items = sorted(current_dir.keys())
            if index < 0 or index >= len(items):
                print("Invalid index")
                return
                
            item = items[index]
            if not isinstance(current_dir[item], dict) or 'id' in current_dir[item]:
                print("Not a directory")
                return
                
            self.current_path.append(item)
            self.do_ls("")
        except ValueError:
            print("Please enter a valid number")

    def do_stream(self, arg):
        """Show stream link for a file: stream <file_index>"""
        if not self.current_torrent:
            print("No torrent selected")
            return
            
        current_dir = self.tree
        for part in self.current_path:
            current_dir = current_dir.get(part, {})
            
        try:
            index = int(arg) - 1
            items = sorted(current_dir.keys())
            if index < 0 or index >= len(items):
                print("Invalid index")
                return
                
            item = items[index]
            if isinstance(current_dir[item], dict) and 'id' not in current_dir[item]:
                print("Not a file")
                return
                
            file_id = current_dir[item]['id']
            stream_url = urljoin(
                self.base_url, 
                f"/play/{self.current_torrent['hash']}/{file_id}"
            )
            print(f"\nStream URL for {item}:")
            print(stream_url)
            print("\nYou can use this URL in media players like VLC or MPV")
            
        except ValueError:
            print("Please enter a valid number")

    def do_delete(self, arg):
        """Delete a torrent: delete <torrent_index>"""
        try:
            index = int(arg) - 1
            data = {"action": "list"}
            torrents = self._api_request("/torrents", json_data=data)
            if not torrents or index < 0 or index >= len(torrents):
                print("Invalid torrent index")
                return
                
            torrent = torrents[index]
            confirm = input(f"Delete '{torrent['title']}'? (y/n): ")
            if confirm.lower() != 'y':
                return
                
            del_data = {
                "action": "rem",
                "hash": torrent['hash']
            }
            result = self._api_request("/torrents", json_data=del_data)
            if result:
                print("Torrent deleted successfully!")
                if self.current_torrent and self.current_torrent['hash'] == torrent['hash']:
                    self.current_torrent = None
                    self.current_path = []
        except ValueError:
            print("Please enter a valid number")

    def do_exit(self, arg):
        """Exit the program"""
        print("Goodbye!")
        return True

def main():
    if len(sys.argv) == 2:
        base_url = sys.argv[1]
    else:
        base_url = input("Enter server URL (default: http://localhost:8090): ") or "http://localhost:8090"
    base_url = base_url.strip()
    if not (base_url.startswith("http://") or base_url.startswith("https://")):
        base_url = "http://" + base_url

    print(f"Connecting to Torrserver at {base_url}")
    print("Type 'help' for available commands\n")
    TorrServerCLI(base_url).cmdloop()

if __name__ == "__main__":
    main()