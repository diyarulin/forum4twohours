document.addEventListener('DOMContentLoaded', function() {
    const setupEventSource = () => {
        const eventSource = new EventSource('/notifications/stream');
        
        eventSource.onmessage = function(e) {
            fetch('/notifications')
                .then(response => response.text())
                .then(html => {
                    const parser = new DOMParser();
                    const doc = parser.parseFromString(html, 'text/html');
                    const newContent = doc.getElementById('notifications-container').innerHTML;
                    document.getElementById('notifications-container').innerHTML = newContent;
                });
        };
        
        eventSource.onerror = function(err) {
            console.error('EventSource failed:', err);
            eventSource.close();
            // Attempt reconnect after 5 seconds
            setTimeout(setupEventSource, 5000);
        };
    };

    setupEventSource();
});