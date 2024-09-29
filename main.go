package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = `
<!DOCTYPE html>
<html>
<head>
    <title>Emoji Gallery</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        h1 {
            text-align: center;
        }
        .search-bar {
            margin-bottom: 20px;
            text-align: center;
        }
        .filter-bar {
            margin-bottom: 20px;
            text-align: center;
        }
        input[type="text"] {
            width: 300px;
            padding: 10px;
            font-size: 16px;
        }
        select {
            padding: 10px;
            font-size: 16px;
        }
        .gallery {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
        }
        .emoji-card {
            border: 1px solid #ddd;
            border-radius: 8px;
            margin: 10px;
            padding: 15px;
            width: 120px;
            text-align: center;
            box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.1);
        }
        .emoji-code {
            font-size: 2em;
            cursor: pointer;
        }
        .emoji-name {
            font-weight: bold;
            margin-top: 10px;
        }
        .hidden {
            display: none;
        }
        .notification {
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background-color: #4CAF50;
            color: white;
            padding: 10px 20px;
            border-radius: 5px;
            opacity: 1;
            transition: opacity 0.5s ease;
            z-index: 1000;
        }
        .notification.fade-out {
            opacity: 0;
        }
    </style>
    <script>
        // Function to filter the emoji gallery based on the search input and group filter
        function filterEmojis() {
            var input = document.getElementById("searchInput").value.toLowerCase();
            var selectedGroup = document.getElementById("groupFilter").value.toLowerCase();
            var emojiCards = document.getElementsByClassName("emoji-card");

            for (var i = 0; i < emojiCards.length; i++) {
                var emojiName = emojiCards[i].getElementsByClassName("emoji-name")[0].innerText.toLowerCase();
                var emojiCategory = emojiCards[i].getElementsByClassName("emoji-category")[0].innerText.toLowerCase();
                var emojiGroup = emojiCards[i].getElementsByClassName("emoji-group")[0].innerText.toLowerCase();

                var nameMatches = emojiName.includes(input) || emojiCategory.includes(input);
                var groupMatches = selectedGroup === "all" || emojiGroup === selectedGroup;

                if (nameMatches && groupMatches) {
                    emojiCards[i].classList.remove("hidden");
                } else {
                    emojiCards[i].classList.add("hidden");
                }
            }
        }

        // Function to copy the emoji to the clipboard when clicked
        function copyEmoji(emojiCode) {
            navigator.clipboard.writeText(emojiCode).then(function() {
                showNotification("Emoji copied to clipboard: " + emojiCode);
            }, function(err) {
                console.error("Could not copy emoji: ", err);
            });
        }

        // Function to show a fade-out notification
        function showNotification(message) {
            var notification = document.createElement("div");
            notification.className = "notification";
            notification.innerText = message;
            document.body.appendChild(notification);

            // Fade out after 3 seconds
            setTimeout(function() {
                notification.classList.add("fade-out");
                // Remove the element from the DOM after the fade-out transition
                setTimeout(function() {
                    document.body.removeChild(notification);
                }, 500); // Wait for the fade-out transition to complete
            }, 2000); // Show notification for 2 seconds
        }

        // Attach click event listeners to all emoji code elements once the DOM is fully loaded
        document.addEventListener("DOMContentLoaded", function() {
            var emojiCodes = document.getElementsByClassName("emoji-code");

            for (var i = 0; i < emojiCodes.length; i++) {
                // Add click event listener to each emoji
                emojiCodes[i].addEventListener("click", function() {
                    var emoji = this.innerHTML.trim(); // Get the emoji (the raw HTML code)
                    copyEmoji(emoji);
                });
            }
        });
    </script>
</head>
<body>
    <h1>Emoji Gallery</h1>

    <!-- Search Bar -->
    <div class="search-bar">
        <input type="text" id="searchInput" placeholder="Search emojis..." onkeyup="filterEmojis()">
    </div>

    <!-- Group Filter Dropdown -->
    <div class="filter-bar">
        <label for="groupFilter">Filter by group:</label>
        <select id="groupFilter" onchange="filterEmojis()">
            <option value="all">All</option>
            {{range .Groups}}
                <option value="{{.}}">{{.}}</option>
            {{end}}
        </select>
    </div>

    <!-- Emoji Gallery -->
    <div class="gallery">
    {{range .Data}}
        <div class="emoji-card">
            <div class="emoji-code">{{index .HtmlCode 0}}</div> <!-- Emoji will be copied when clicked -->
            <div class="emoji-name">{{.Name}}</div>
            <div class="emoji-category">{{.Category}}</div>
            <div class="emoji-group">{{.Group}}</div>
        </div>
    {{end}}
    </div>

</body>
</html>
`

type Emoji struct {
	Name     string          `json:"name"`
	Category string          `json:"category"`
	Group    string          `json:"group"`
	HtmlCode []template.HTML `json:"htmlCode"`
	Unicode  []string        `json:"unicode"`
}

type ApiResponse struct {
	Status string  `json:"status"`
	Data   []Emoji `json:"data"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	url := "https://emojihub.yurace.pro/api/all"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	var apiResponse []Emoji
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		return
	}

	groupMap := make(map[string]bool)
	for _, emoji := range apiResponse {
		groupMap[emoji.Group] = true
	}

	var groups []string
	for group := range groupMap {
		groups = append(groups, group)
	}

	type TemplateData struct {
		Data   []Emoji
		Groups []string
	}

	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	t.Execute(w, TemplateData{Data: apiResponse, Groups: groups})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
