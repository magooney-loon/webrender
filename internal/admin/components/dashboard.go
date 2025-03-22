package components

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/magooney-loon/webrender/pkg/component"
)

const (
	dashboardTemplate = `
	<div id="{{.ID}}" class="bg-transparent rounded-lg p-6 w-full max-w-7xl" data-state='{{.State.ToJSON}}' data-component-type="AdminDashboard">
		<div class="flex justify-between items-center mb-6">
			<h1 class="text-2xl font-semibold text-white flex items-center">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mr-2 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
					<path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
				</svg>
				Admin Dashboard
			</h1>
			<div class="text-sm text-vercel-gray-400 flex items-center">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-vercel-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Last updated: <span data-bind="lastUpdated" class="ml-1 text-vercel-gray-300">{{.State.Get "lastUpdated"}}</span>
			</div>
		</div>
		
		<!-- Notification area for feedback -->
		<div id="{{.ID}}-notification" class="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50 transition-all duration-300 opacity-0 translate-y-4 hidden">
			<div class="p-4 rounded-md bg-vercel-gray-800 border border-vercel-accent-400 shadow-lg max-w-md">
				<p class="text-white font-medium flex items-center" data-bind="notification">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
					</svg>
					{{.State.Get "notification"}}
				</p>
			</div>
		</div>
		
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
			<!-- Stats Cards -->
			<div class="vercel-card p-5">
				<h3 class="text-sm font-medium text-vercel-gray-400 mb-1">Users</h3>
				<p class="text-4xl font-mono font-semibold text-white mb-2" data-bind="users">{{.State.Get "users"}}</p>
				<div class="flex items-center text-sm">
					<span data-bind="userTrendIcon" class="mr-1" style="color: {{.State.Get "userTrendColor"}}">{{.State.Get "userTrendIcon"}}</span>
					<span data-bind="userTrend" style="color: {{.State.Get "userTrendColor"}}">{{.State.Get "userTrend"}}</span>
					<span class="text-vercel-gray-400 ml-1">% from last week</span>
				</div>
				<div class="h-1 w-full bg-vercel-gray-700 mt-3 rounded-full overflow-hidden">
					<div class="h-1 bg-blue-500 progress-bar rounded-full" style="width: {{.State.Get "usersPercentage"}}%"></div>
				</div>
			</div>
			
			<div class="vercel-card p-5">
				<h3 class="text-sm font-medium text-vercel-gray-400 mb-1">Active Sessions</h3>
				<p class="text-4xl font-mono font-semibold text-white mb-2" data-bind="sessions">{{.State.Get "sessions"}}</p>
				<div class="flex items-center text-sm">
					<span data-bind="sessionTrendIcon" class="mr-1" style="color: {{.State.Get "sessionTrendColor"}}">{{.State.Get "sessionTrendIcon"}}</span>
					<span data-bind="sessionTrend" style="color: {{.State.Get "sessionTrendColor"}}">{{.State.Get "sessionTrend"}}</span>
					<span class="text-vercel-gray-400 ml-1">% from yesterday</span>
				</div>
				<div class="h-1 w-full bg-vercel-gray-700 mt-3 rounded-full overflow-hidden">
					<div class="h-1 bg-green-500 progress-bar rounded-full" style="width: {{.State.Get "sessionsPercentage"}}%"></div>
				</div>
			</div>
			
			<div class="vercel-card p-5">
				<h3 class="text-sm font-medium text-vercel-gray-400 mb-1">Server Load</h3>
				<p class="text-4xl font-mono font-semibold text-white mb-2" data-bind="serverLoad">{{.State.Get "serverLoad"}}</p>
				<div class="flex items-center text-sm">
					<span data-bind="loadTrendIcon" class="mr-1" style="color: {{.State.Get "loadTrendColor"}}">{{.State.Get "loadTrendIcon"}}</span>
					<span data-bind="loadTrend" style="color: {{.State.Get "loadTrendColor"}}">{{.State.Get "loadTrend"}}</span>
					<span class="text-vercel-gray-400 ml-1">% from average</span>
				</div>
				<div class="h-1 w-full bg-vercel-gray-700 mt-3 rounded-full overflow-hidden">
					<div class="h-1 bg-purple-500 progress-bar rounded-full" style="width: {{.State.Get "loadPercentage"}}%"></div>
				</div>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
			<!-- Traffic Graph -->
			<div class="vercel-card p-5">
				<h3 class="text-sm font-medium text-vercel-gray-400 mb-4 flex items-center">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
						<path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
					</svg>
					Current Traffic
				</h3>
				<div class="h-32 flex items-end space-x-1">
					{{range $i, $height := .State.Get "trafficGraph"}}
					<div class="w-full bg-indigo-{{$height}} h-{{$height}} rounded-t-sm"></div>
					{{end}}
				</div>
				<div class="flex justify-between mt-2 text-xs text-vercel-gray-500">
					<span>5m ago</span>
					<span>4m ago</span>
					<span>3m ago</span>
					<span>2m ago</span>
					<span>1m ago</span>
					<span>Now</span>
				</div>
			</div>

			<!-- Recent Events -->
			<div class="vercel-card p-5">
				<h3 class="text-sm font-medium text-vercel-gray-400 mb-4 flex items-center">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
					</svg>
					Recent Events
				</h3>
				<ul class="space-y-2 text-sm">
					{{range $event := .State.Get "recentEvents"}}
					<li class="py-2 border-b border-vercel-gray-700 flex items-center">
						<span class="text-{{$event.color}}-400 mr-2">{{$event.icon}}</span>
						<span class="text-white">{{$event.text}}</span>
						<span class="text-vercel-gray-500 ml-auto font-mono">{{$event.time}}</span>
					</li>
					{{end}}
				</ul>
			</div>
		</div>
		
		<div class="mb-6">
			<div class="flex items-center mb-4">
				<div class="w-3 h-3 rounded-full bg-vercel-accent-400 mr-2"></div>
				<h2 class="text-xl font-semibold text-white">System Status</h2>
			</div>
			<div class="vercel-card p-0 overflow-hidden">
				<div class="flex items-center justify-between p-5 border-b border-vercel-gray-800">
					<div class="flex items-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3 text-vercel-gray-400" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M2 5a2 2 0 012-2h12a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V5zm3.293 1.293a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 01-1.414-1.414L7.586 10 5.293 7.707a1 1 0 010-1.414zM11 12a1 1 0 100 2h3a1 1 0 100-2h-3z" clip-rule="evenodd" />
						</svg>
						<span class="text-white text-sm font-medium">WebSocket Service</span>
					</div>
					<span class="px-3 py-1 rounded-full text-xs font-medium" 
					      data-bind="wsStatus"
					      style="background-color: {{.State.Get "wsStatusColor"}}; color: {{.State.Get "wsStatusTextColor"}};">
						{{.State.Get "wsStatus"}}
					</span>
				</div>
				<div class="flex items-center justify-between p-5 border-b border-vercel-gray-800">
					<div class="flex items-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3 text-vercel-gray-400" viewBox="0 0 20 20" fill="currentColor">
							<path d="M3 12v3c0 1.657 3.134 3 7 3s7-1.343 7-3v-3c0 1.657-3.134 3-7 3s-7-1.343-7-3z" />
							<path d="M3 7v3c0 1.657 3.134 3 7 3s7-1.343 7-3V7c0 1.657-3.134 3-7 3S3 8.657 3 7z" />
							<path d="M17 5c0 1.657-3.134 3-7 3S3 6.657 3 5s3.134-3 7-3 7 1.343 7 3z" />
						</svg>
						<span class="text-white text-sm font-medium">Database</span>
					</div>
					<span class="px-3 py-1 rounded-full text-xs font-medium" 
					      data-bind="dbStatus"
					      style="background-color: {{.State.Get "dbStatusColor"}}; color: {{.State.Get "dbStatusTextColor"}};">
						{{.State.Get "dbStatus"}}
					</span>
				</div>
				<div class="flex items-center justify-between p-5">
					<div class="flex items-center">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3 text-vercel-gray-400" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
						</svg>
						<span class="text-white text-sm font-medium">Cache Service</span>
					</div>
					<span class="px-3 py-1 rounded-full text-xs font-medium" 
					      data-bind="cacheStatus"
					      style="background-color: {{.State.Get "cacheStatusColor"}}; color: {{.State.Get "cacheStatusTextColor"}};">
						{{.State.Get "cacheStatus"}}
					</span>
				</div>
			</div>
		</div>
		
		<div>
			<h2 class="text-xl font-semibold text-white mb-4 flex items-center">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-vercel-accent-400" viewBox="0 0 20 20" fill="currentColor">
					<path d="M5 4a1 1 0 00-2 0v7.268a2 2 0 000 3.464V16a1 1 0 102 0v-1.268a2 2 0 000-3.464V4zM11 4a1 1 0 10-2 0v1.268a2 2 0 000 3.464V16a1 1 0 102 0V8.732a2 2 0 000-3.464V4zM16 3a1 1 0 011 1v7.268a2 2 0 010 3.464V16a1 1 0 11-2 0v-1.268a2 2 0 010-3.464V4a1 1 0 011-1z" />
				</svg>
				Quick Actions
			</h2>
			<div class="flex flex-wrap gap-4">
				<button onclick="AdminDashboard.refreshStats('{{.ID}}')" 
				        class="vercel-btn vercel-btn-primary">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd" />
					</svg>
					Refresh Statistics
				</button>
				<button onclick="AdminDashboard.clearCache('{{.ID}}')" 
				        class="vercel-btn vercel-btn-secondary">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M4 2a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1V3a1 1 0 00-1-1H4zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1V3a1 1 0 00-1-1H7zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1V3a1 1 0 00-1-1h-1z" clip-rule="evenodd" />
						<path d="M9 6a1 1 0 011-1h1a1 1 0 110 2H9a1 1 0 01-1-1z" />
						<path fill-rule="evenodd" d="M2 9a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H2zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H5zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H8z" clip-rule="evenodd" />
						<path d="M9 11a1 1 0 011-1h1a1 1 0 110 2H9a1 1 0 01-1-1z" />
						<path fill-rule="evenodd" d="M2 13a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H2zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H5zm3 0a1 1 0 00-1 1v1a1 1 0 001 1h1a1 1 0 001-1v-1a1 1 0 00-1-1H8z" clip-rule="evenodd" />
					</svg>
					Clear Cache
				</button>
				<button onclick="AdminDashboard.checkSystem('{{.ID}}')" 
				        class="vercel-btn vercel-btn-secondary">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 1.944A11.954 11.954 0 012.166 5C2.056 5.649 2 6.319 2 7c0 5.225 3.34 9.67 8 11.317C14.66 16.67 18 12.225 18 7c0-.682-.057-1.35-.166-2.001A11.954 11.954 0 0110 1.944zM11 14a1 1 0 01-2 0v-2a1 1 0 012 0v2zm0-6a1 1 0 01-2 0V6a1 1 0 012 0v2z" clip-rule="evenodd" />
					</svg>
					Check System Health
				</button>
			</div>
		</div>
	</div>
	`

	dashboardStyles = `
	.h-1 { height: 0.25rem; }
	.h-2 { height: 0.5rem; }
	.h-3 { height: 0.75rem; }
	.h-4 { height: 1rem; }
	.h-5 { height: 1.25rem; }
	.h-6 { height: 1.5rem; }
	.h-8 { height: 2rem; }
	.h-10 { height: 2.5rem; }
	.h-12 { height: 3rem; }
	.h-16 { height: 4rem; }
	.h-20 { height: 5rem; }
	.h-24 { height: 6rem; }
	.h-28 { height: 7rem; }
	.h-32 { height: 8rem; }
	.bg-indigo-1 { background-color: rgba(99, 102, 241, 0.1); }
	.bg-indigo-2 { background-color: rgba(99, 102, 241, 0.2); }
	.bg-indigo-3 { background-color: rgba(99, 102, 241, 0.3); }
	.bg-indigo-4 { background-color: rgba(99, 102, 241, 0.4); }
	.bg-indigo-5 { background-color: rgba(99, 102, 241, 0.5); }
	.bg-indigo-6 { background-color: rgba(99, 102, 241, 0.6); }
	.bg-indigo-7 { background-color: rgba(99, 102, 241, 0.7); }
	.bg-indigo-8 { background-color: rgba(99, 102, 241, 0.8); }
	.bg-indigo-9 { background-color: rgba(99, 102, 241, 0.9); }
	.bg-indigo-10 { background-color: rgba(99, 102, 241, 1); }
	`

	dashboardScript = `
	// AdminDashboard component handler
	const AdminDashboard = {
		// Show notification message
		showNotification(componentId, message) {
			const component = document.getElementById(componentId);
			const notificationArea = document.getElementById(componentId + '-notification');
			const messageEl = notificationArea.querySelector('[data-bind="notification"]');
			
			// Set message
			if (messageEl.childNodes.length > 1) {
				messageEl.childNodes[messageEl.childNodes.length - 1].textContent = message;
			} else {
				messageEl.textContent = message;
			}
			
			// Show the notification
			notificationArea.classList.remove('hidden');
			setTimeout(() => {
				notificationArea.classList.remove('opacity-0');
				notificationArea.classList.remove('translate-y-4');
			}, 10);
			
			// Hide after 3 seconds
			setTimeout(() => {
				notificationArea.classList.add('opacity-0');
				notificationArea.classList.add('translate-y-4');
				setTimeout(() => notificationArea.classList.add('hidden'), 300);
			}, 3000);
		},
		
		refreshStats(componentId) {
			console.log("Refreshing dashboard stats...");
			
			// Show notification
			this.showNotification(componentId, "Refreshing statistics...");
			
			// Send update to server
			WSManager.sendAction(componentId, "refreshStats", {});
		},
		
		clearCache(componentId) {
			console.log("Clearing system cache...");
			
			// Show notification
			this.showNotification(componentId, "Clearing cache...");
			
			// Send update to server
			WSManager.sendAction(componentId, "clearCache", {});
		},
		
		checkSystem(componentId) {
			console.log("Running system health check...");
			
			// Show notification
			this.showNotification(componentId, "Running system health check...");
			
			// Send update to server
			WSManager.sendAction(componentId, "checkSystem", {});
		},
		
		// Handle state updates from the server
		updateStats(componentId, data) {
			const component = document.getElementById(componentId);
			
			// Update UI elements based on data
			for (const [key, value] of Object.entries(data)) {
				const element = component.querySelector('[data-bind="' + key + '"]');
				if (element) {
					if (key.endsWith("Color")) {
						element.style.backgroundColor = value;
					} else if (key.endsWith("TrendColor")) {
						element.style.color = value;
					} else if (key.includes("Percentage")) {
						element.style.width = value + "%";
					} else {
						element.textContent = value;
					}
				}
			}
			
			// If there's a notification message, show it
			if (data.notification) {
				this.showNotification(componentId, data.notification);
			}
		}
	};
	`
)

// Color constants for status indicators
const (
	// Status colors
	colorHealthy = "rgba(46, 204, 113, 0.15)" // Light green background with transparency
	colorWarning = "rgba(241, 196, 15, 0.15)" // Light yellow background with transparency
	colorError   = "rgba(231, 76, 60, 0.15)"  // Light red background with transparency

	// Text colors for status indicators
	textColorHealthy = "#2ecc71" // Green text
	textColorWarning = "#f1c40f" // Yellow text
	textColorError   = "#e74c3c" // Red text

	// Trend colors
	colorPositive = "#4ade80" // Green for positive trends
	colorNeutral  = "#a1a1aa" // Gray for neutral trends
	colorNegative = "#f87171" // Red for negative trends

	// Icons
	iconUp      = "↑"
	iconDown    = "↓"
	iconNeutral = "→"
)

// TrafficPattern simulates real web traffic patterns throughout the day
type TrafficPattern struct {
	baseUsers        int
	baseSessionsRate float64
	baseServerLoad   int
	hourlyPatterns   map[int]float64
	nextUpdateTime   time.Time
	lastValues       map[string]interface{}
	mutex            sync.Mutex
	stopChan         chan struct{}
}

// NewTrafficPattern creates a realistic traffic pattern simulator
func NewTrafficPattern() *TrafficPattern {
	// Hourly multipliers representing a typical day's traffic
	// 0-23 hours with multipliers (1.0 is baseline)
	hourlyPatterns := map[int]float64{
		0: 0.5, 1: 0.3, 2: 0.2, 3: 0.15, 4: 0.15, 5: 0.3,
		6: 0.7, 7: 1.1, 8: 1.5, 9: 1.8, 10: 1.9, 11: 2.0,
		12: 2.1, 13: 2.0, 14: 1.9, 15: 1.8, 16: 1.7, 17: 1.5,
		18: 1.3, 19: 1.1, 20: 0.9, 21: 0.8, 22: 0.7, 23: 0.6,
	}

	return &TrafficPattern{
		baseUsers:        5000,
		baseSessionsRate: 0.1, // 10% of users are active in sessions
		baseServerLoad:   15,  // 15% baseline server load
		hourlyPatterns:   hourlyPatterns,
		nextUpdateTime:   time.Now(),
		lastValues:       make(map[string]interface{}),
		mutex:            sync.Mutex{},
		stopChan:         make(chan struct{}),
	}
}

// GetCurrentMultiplier returns traffic multiplier based on current time
func (tp *TrafficPattern) GetCurrentMultiplier() float64 {
	hour := time.Now().Hour()
	// Get base multiplier for the hour
	multiplier := tp.hourlyPatterns[hour]

	// Add some randomness (±10%)
	randomFactor := 0.9 + (rand.Float64() * 0.2)
	return multiplier * randomFactor
}

// GenerateTrafficData generates realistic traffic data
func (tp *TrafficPattern) GenerateTrafficData() map[string]interface{} {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	data := make(map[string]interface{})

	// Get traffic multiplier based on time of day
	multiplier := tp.GetCurrentMultiplier()

	// Calculate user count with some randomness
	usersBase := float64(tp.baseUsers) * multiplier
	usersRandom := usersBase * (0.97 + rand.Float64()*0.06) // ±3% randomness
	users := int(usersRandom)

	// Ensure users is never zero
	if users < 1 {
		users = int(usersBase)
		if users < 1 {
			users = tp.baseUsers / 2
		}
	}

	// Session count (active users)
	sessionRate := tp.baseSessionsRate * (0.9 + rand.Float64()*0.2) // Vary session rate
	sessions := int(float64(users) * sessionRate)

	// Ensure sessions is never zero
	if sessions < 1 {
		sessions = int(float64(users) * tp.baseSessionsRate)
		if sessions < 1 {
			sessions = 1
		}
	}

	// Server load (correlated with users but not directly proportional)
	serverLoadBase := tp.baseServerLoad + int(45.0*(multiplier-0.5)/(2.0-0.5))
	serverLoadRandom := serverLoadBase + rand.Intn(11) - 5 // ±5%
	if serverLoadRandom < 10 {
		serverLoadRandom = 10
	} else if serverLoadRandom > 95 {
		serverLoadRandom = 95
	}

	// Calculate trends compared to last values
	userTrend := 0.0
	sessionTrend := 0.0
	loadTrend := 0.0

	if lastUsers, ok := tp.lastValues["users"].(int); ok && lastUsers > 0 {
		userTrend = float64(users-lastUsers) / float64(lastUsers) * 100
	}

	if lastSessions, ok := tp.lastValues["sessions"].(int); ok && lastSessions > 0 {
		sessionTrend = float64(sessions-lastSessions) / float64(lastSessions) * 100
	}

	if lastLoad, ok := tp.lastValues["serverLoadValue"].(int); ok && lastLoad > 0 {
		loadTrend = float64(serverLoadRandom-lastLoad) / float64(lastLoad) * 100
	}

	// Round trends to 1 decimal place
	userTrend = float64(int(userTrend*10)) / 10
	sessionTrend = float64(int(sessionTrend*10)) / 10
	loadTrend = float64(int(loadTrend*10)) / 10

	// Set trend colors and icons
	userTrendColor := colorNeutral
	userTrendIcon := iconNeutral
	if userTrend > 0.3 {
		userTrendColor = colorPositive
		userTrendIcon = iconUp
	} else if userTrend < -0.3 {
		userTrendColor = colorNegative
		userTrendIcon = iconDown
	}

	sessionTrendColor := colorNeutral
	sessionTrendIcon := iconNeutral
	if sessionTrend > 0.3 {
		sessionTrendColor = colorPositive
		sessionTrendIcon = iconUp
	} else if sessionTrend < -0.3 {
		sessionTrendColor = colorNegative
		sessionTrendIcon = iconDown
	}

	loadTrendColor := colorNeutral
	loadTrendIcon := iconNeutral
	if loadTrend > 0.3 {
		loadTrendColor = colorNegative // Higher load is negative
		loadTrendIcon = iconUp
	} else if loadTrend < -0.3 {
		loadTrendColor = colorPositive // Lower load is positive
		loadTrendIcon = iconDown
	}

	// Store the current values for next comparison
	tp.lastValues["users"] = users
	tp.lastValues["sessions"] = sessions
	tp.lastValues["serverLoadValue"] = serverLoadRandom

	// Traffic graph - generate heights for 12 bars
	trafficGraph := make([]int, 12)
	for i := 0; i < 12; i++ {
		// Create a pattern with some randomness
		height := 1 + int(9*multiplier*(0.7+rand.Float64()*0.6))
		if height > 10 {
			height = 10
		}
		trafficGraph[i] = height
	}

	// Recent events - create simulated events
	recentEvents := []map[string]string{
		generateRandomEvent(),
		generateRandomEvent(),
		generateRandomEvent(),
		generateRandomEvent(),
		generateRandomEvent(),
	}

	// Fill the data map
	data["users"] = users
	data["userTrend"] = userTrend
	data["userTrendColor"] = userTrendColor
	data["userTrendIcon"] = userTrendIcon
	data["usersPercentage"] = 50 + int(userTrend) // Visual indicator for trend

	data["sessions"] = sessions
	data["sessionTrend"] = sessionTrend
	data["sessionTrendColor"] = sessionTrendColor
	data["sessionTrendIcon"] = sessionTrendIcon
	data["sessionsPercentage"] = 50 + int(sessionTrend)

	data["serverLoad"] = fmt.Sprintf("%d%%", serverLoadRandom)
	data["loadTrend"] = loadTrend
	data["loadTrendColor"] = loadTrendColor
	data["loadTrendIcon"] = loadTrendIcon
	data["loadPercentage"] = serverLoadRandom

	data["trafficGraph"] = trafficGraph
	data["recentEvents"] = recentEvents

	return data
}

// EventTemplates for random event generation
var eventTemplates = []struct {
	text  string
	color string
	icon  string
}{
	{"User %s logged in", "green", "●"},
	{"File upload completed: %s", "blue", "↑"},
	{"Database query optimized: %s", "purple", "⚡"},
	{"API request from %s", "blue", "→"},
	{"Cache invalidated for %s", "amber", "♻"},
	{"Background job completed: %s", "green", "✓"},
	{"New user registered: %s", "green", "+"},
	{"Password reset requested for %s", "amber", "↻"},
	{"Config updated: %s", "blue", "⚙"},
	{"Session timeout for %s", "red", "✗"},
	{"Error rate increased for %s", "red", "!"},
	{"Memory usage decreased by %s", "green", "↓"},
	{"CDN cache refreshed for %s", "blue", "↻"},
	{"Backup completed: %s", "green", "✓"},
	{"Auth failed for %s", "red", "✗"},
}

// generateRandomEvent creates a random event for the dashboard
func generateRandomEvent() map[string]string {
	template := eventTemplates[rand.Intn(len(eventTemplates))]

	// Random parameters for the event text
	params := []string{
		fmt.Sprintf("user%d", rand.Intn(9999)),
		fmt.Sprintf("task-%d", rand.Intn(1000)),
		fmt.Sprintf("profile-%d", rand.Intn(5000)),
		fmt.Sprintf("product-%d", rand.Intn(1000)),
		fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255)),
		fmt.Sprintf("api/v1/%s", []string{"users", "products", "orders"}[rand.Intn(3)]),
		fmt.Sprintf("%d%%", 5+rand.Intn(90)),
		fmt.Sprintf("%dKB", 100+rand.Intn(9000)),
	}

	// Create relative timestamp (within the last hour)
	minutes := rand.Intn(60)
	var timeAgo string
	if minutes == 0 {
		timeAgo = "just now"
	} else if minutes == 1 {
		timeAgo = "1 min ago"
	} else {
		timeAgo = fmt.Sprintf("%d mins ago", minutes)
	}

	return map[string]string{
		"text":  fmt.Sprintf(template.text, params[rand.Intn(len(params))]),
		"color": template.color,
		"icon":  template.icon,
		"time":  timeAgo,
	}
}

// StartContinuousUpdates begins sending continuous updates to the component
func (tp *TrafficPattern) StartContinuousUpdates(dashboard *component.Component) {
	// Initial update
	data := tp.GenerateTrafficData()
	for key, value := range data {
		dashboard.State.Set(key, value)
	}
	dashboard.State.Set("lastUpdated", time.Now().Format("Jan 2, 2006 15:04:05"))

	// Store a reference to the current data
	currentUsers := 0
	currentSessions := 0
	currentLoad := 0
	if u, ok := data["users"].(int); ok {
		currentUsers = u
	}
	if s, ok := data["sessions"].(int); ok {
		currentSessions = s
	}
	if l, ok := data["loadPercentage"].(int); ok {
		currentLoad = l
	}

	// Track if we had an event/spike recently to avoid too many
	lastEventTime := time.Now()

	// Start the continuous update goroutine
	go func() {
		// Use two timers:
		// 1. A fast timer for micro-updates (small changes to simulate real-time movement)
		// 2. A slower timer for occasional larger changes and event generation
		fastTicker := time.NewTicker(200 * time.Millisecond)
		slowTicker := time.NewTicker(3 * time.Second)
		defer fastTicker.Stop()
		defer slowTicker.Stop()

		// Track traffic graph data with each small update
		trafficData := []int{4, 5, 4, 6, 5, 7, 6, 8, 7, 6, 7, 8}
		eventLog := make([]map[string]string, 5)
		for i := 0; i < 5; i++ {
			eventLog[i] = generateRandomEvent()
		}

		for {
			select {
			case <-tp.stopChan:
				return
			case <-fastTicker.C:
				// Small incremental updates (micro-changes to simulate real-time)
				tp.mutex.Lock()

				// Micro-change for users (±0.1%)
				microChangeUsers := float64(currentUsers) * (0.999 + rand.Float64()*0.002)
				newUsers := int(microChangeUsers)
				if newUsers < 1 {
					newUsers = tp.baseUsers / 2
					if newUsers < 1 {
						newUsers = 100 // Fallback minimum
					}
				}

				// Micro-change for sessions
				sessionRate := tp.baseSessionsRate * (0.99 + rand.Float64()*0.02)
				newSessions := int(float64(newUsers) * sessionRate)
				if newSessions < 1 {
					newSessions = 1 // Ensure at least one active session
				}

				// Micro-change for server load (±0.2%)
				loadChange := currentLoad + rand.Intn(3) - 1
				if loadChange < 10 {
					loadChange = 10
				} else if loadChange > 95 {
					loadChange = 95
				}

				// Calculate trends
				userTrend := 0.0
				sessionTrend := 0.0
				loadTrend := 0.0

				if currentUsers > 0 {
					userTrend = float64(int(((float64(newUsers)-float64(currentUsers))/float64(currentUsers)*100)*10)) / 10
				}

				if currentSessions > 0 {
					sessionTrend = float64(int(((float64(newSessions)-float64(currentSessions))/float64(currentSessions)*100)*10)) / 10
				}

				if currentLoad > 0 {
					loadTrend = float64(int(((float64(loadChange)-float64(currentLoad))/float64(currentLoad)*100)*10)) / 10
				}

				// Update trend indicators
				userTrendColor := colorNeutral
				userTrendIcon := iconNeutral
				if userTrend > 0.1 {
					userTrendColor = colorPositive
					userTrendIcon = iconUp
				} else if userTrend < -0.1 {
					userTrendColor = colorNegative
					userTrendIcon = iconDown
				}

				sessionTrendColor := colorNeutral
				sessionTrendIcon := iconNeutral
				if sessionTrend > 0.1 {
					sessionTrendColor = colorPositive
					sessionTrendIcon = iconUp
				} else if sessionTrend < -0.1 {
					sessionTrendColor = colorNegative
					sessionTrendIcon = iconDown
				}

				loadTrendColor := colorNeutral
				loadTrendIcon := iconNeutral
				if loadTrend > 0.1 {
					loadTrendColor = colorNegative
					loadTrendIcon = iconUp
				} else if loadTrend < -0.1 {
					loadTrendColor = colorPositive
					loadTrendIcon = iconDown
				}

				// Update traffic graph - shift values and add new one
				newHeight := 1 + int(9*tp.GetCurrentMultiplier()*(0.7+rand.Float64()*0.6))
				if newHeight > 10 {
					newHeight = 10
				}
				for i := 0; i < len(trafficData)-1; i++ {
					trafficData[i] = trafficData[i+1]
				}
				trafficData[len(trafficData)-1] = newHeight

				// Random chance for a minor traffic spike or dip (1% chance)
				if rand.Float64() < 0.01 {
					// Create a small spike or dip
					spikeFactor := 1.0 + (rand.Float64()*0.1 - 0.05) // ±5% change
					newUsers = int(float64(newUsers) * spikeFactor)
					newSessions = int(float64(newSessions) * spikeFactor)

					// Reflect in the traffic graph
					spikeHeight := trafficData[len(trafficData)-1]
					if spikeFactor > 1 {
						spikeHeight += 2
						if spikeHeight > 10 {
							spikeHeight = 10
						}
					} else {
						spikeHeight -= 2
						if spikeHeight < 1 {
							spikeHeight = 1
						}
					}
					trafficData[len(trafficData)-1] = spikeHeight
				}

				// Update data for the UI
				dashboard.State.Set("users", newUsers)
				dashboard.State.Set("userTrend", userTrend)
				dashboard.State.Set("userTrendColor", userTrendColor)
				dashboard.State.Set("userTrendIcon", userTrendIcon)
				dashboard.State.Set("sessions", newSessions)
				dashboard.State.Set("sessionTrend", sessionTrend)
				dashboard.State.Set("sessionTrendColor", sessionTrendColor)
				dashboard.State.Set("sessionTrendIcon", sessionTrendIcon)
				dashboard.State.Set("loadPercentage", loadChange)
				dashboard.State.Set("loadTrend", loadTrend)
				dashboard.State.Set("loadTrendColor", loadTrendColor)
				dashboard.State.Set("loadTrendIcon", loadTrendIcon)

				// Update our tracking variables for next time
				currentUsers = newUsers
				currentSessions = newSessions
				currentLoad = loadChange

				// Save to lastValues for trend calculations
				tp.lastValues["users"] = newUsers
				tp.lastValues["sessions"] = newSessions
				tp.lastValues["serverLoadValue"] = loadChange

				tp.mutex.Unlock()

			case <-slowTicker.C:
				// Larger changes, including possible traffic spikes and new events
				tp.mutex.Lock()

				// Every ~3 seconds, update the timestamp
				dashboard.State.Set("lastUpdated", time.Now().Format("Jan 2, 2006 15:04:05"))

				// Add a new random event at the top of the list
				newEvent := generateRandomEvent()
				for i := len(eventLog) - 1; i > 0; i-- {
					eventLog[i] = eventLog[i-1]
				}
				eventLog[0] = newEvent
				dashboard.State.Set("recentEvents", eventLog)

				// Random chance (15%) of a significant traffic event if it's been at least 15 seconds
				if rand.Float64() < 0.15 && time.Since(lastEventTime) > 15*time.Second {
					lastEventTime = time.Now()

					// Decide between spike or dip
					var eventType string
					var changeFactor float64

					if rand.Float64() < 0.6 {
						// Traffic spike (60% chance)
						eventType = "spike"
						changeFactor = 1.05 + rand.Float64()*0.2 // 5-25% increase
					} else {
						// Traffic dip (40% chance)
						eventType = "dip"
						changeFactor = 0.8 + rand.Float64()*0.15 // 5-25% decrease
					}

					// Apply the change
					currentUsers = int(float64(currentUsers) * changeFactor)
					currentSessions = int(float64(currentSessions) * changeFactor)

					// Create an appropriate notification
					var notification string
					if eventType == "spike" {
						notification = "Traffic spike detected! " + fmt.Sprintf("%.1f", (changeFactor-1.0)*100) + "% increase in user activity."
						// Also affect server load
						currentLoad += int((changeFactor - 1.0) * 100)
						if currentLoad > 95 {
							currentLoad = 95
						}
					} else {
						notification = "Traffic decrease detected! " + fmt.Sprintf("%.1f", (1.0-changeFactor)*100) + "% drop in user activity."
						// Also affect server load
						currentLoad -= int((1.0 - changeFactor) * 50)
						if currentLoad < 10 {
							currentLoad = 10
						}
					}

					// Update the UI with notification
					dashboard.State.Set("notification", notification)

					// Add a special event to the event log
					specialEvent := map[string]string{
						"text": notification,
						"time": "just now",
					}

					// Set color and icon based on event type
					if eventType == "spike" {
						specialEvent["color"] = "amber"
						specialEvent["icon"] = "!"
					} else {
						specialEvent["color"] = "blue"
						specialEvent["icon"] = "↓"
					}

					for i := len(eventLog) - 1; i > 0; i-- {
						eventLog[i] = eventLog[i-1]
					}
					eventLog[0] = specialEvent
					dashboard.State.Set("recentEvents", eventLog)

					// Update UI with new values
					dashboard.State.Set("users", currentUsers)
					dashboard.State.Set("sessions", currentSessions)
					dashboard.State.Set("serverLoad", fmt.Sprintf("%d%%", currentLoad))
					dashboard.State.Set("loadPercentage", currentLoad)
				}

				tp.mutex.Unlock()
			}
		}
	}()
}

// StopUpdates stops the continuous update process
func (tp *TrafficPattern) StopUpdates() {
	close(tp.stopChan)
}

// NewAdminDashboard creates a new admin dashboard component
func NewAdminDashboard(id string) *component.Component {
	// Create component with template
	dashboard := component.New(id, "admin-dashboard", dashboardTemplate)

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create traffic pattern simulator
	trafficPattern := NewTrafficPattern()

	// Status indicators
	dashboard.State.Set("wsStatus", "HEALTHY")
	dashboard.State.Set("wsStatusColor", colorHealthy)
	dashboard.State.Set("wsStatusTextColor", textColorHealthy)

	dashboard.State.Set("dbStatus", "HEALTHY")
	dashboard.State.Set("dbStatusColor", colorHealthy)
	dashboard.State.Set("dbStatusTextColor", textColorHealthy)

	dashboard.State.Set("cacheStatus", "HEALTHY")
	dashboard.State.Set("cacheStatusColor", colorHealthy)
	dashboard.State.Set("cacheStatusTextColor", textColorHealthy)

	// Time tracking
	dashboard.State.Set("lastUpdated", time.Now().Format("Jan 2, 2006 15:04:05"))

	// Notification message (initially empty)
	dashboard.State.Set("notification", "")

	// Start continuous updates instead of timed updates
	trafficPattern.StartContinuousUpdates(dashboard)

	// Add method for refreshing stats (will be called via WebSocket)
	dashboard.Methods["refreshStats"] = func(params map[string]interface{}) error {
		// Generate new traffic data with more dramatic changes
		data := trafficPattern.GenerateTrafficData()

		// Apply a "refresh boost" - make changes more noticeable
		if usersVal, ok := data["users"].(int); ok {
			// Boost by ±7-15%
			boostFactor := 0.93 + rand.Float64()*0.22
			data["users"] = int(float64(usersVal) * boostFactor)
		}

		if sessionsVal, ok := data["sessions"].(int); ok {
			// Boost by ±10-20%
			boostFactor := 0.9 + rand.Float64()*0.3
			data["sessions"] = int(float64(sessionsVal) * boostFactor)
		}

		// Update state with new data
		for key, value := range data {
			dashboard.State.Set(key, value)
		}

		dashboard.State.Set("lastUpdated", time.Now().Format("Jan 2, 2006 15:04:05"))
		dashboard.State.Set("notification", "Statistics refreshed successfully!")
		return nil
	}

	dashboard.Methods["clearCache"] = func(params map[string]interface{}) error {
		// Simulate cache clearing
		dashboard.State.Set("cacheStatus", "CLEARING")
		dashboard.State.Set("cacheStatusColor", colorWarning)
		dashboard.State.Set("cacheStatusTextColor", textColorWarning)
		dashboard.State.Set("notification", "Cache clearing in progress...")

		// Simulate a delay for cache clearing
		go func() {
			time.Sleep(1500 * time.Millisecond)

			// Randomly decide if cache clearing was successful or had an issue
			if rand.Float32() > 0.25 {
				// Success
				dashboard.State.Set("cacheStatus", "HEALTHY")
				dashboard.State.Set("cacheStatusColor", colorHealthy)
				dashboard.State.Set("cacheStatusTextColor", textColorHealthy)
				dashboard.State.Set("notification", "Cache cleared successfully!")
			} else {
				// Simulated error
				dashboard.State.Set("cacheStatus", "WARNING")
				dashboard.State.Set("cacheStatusColor", colorWarning)
				dashboard.State.Set("cacheStatusTextColor", textColorWarning)
				dashboard.State.Set("notification", "Cache partially cleared. Some items persisted.")
			}

			// Update stats after cache clear
			data := trafficPattern.GenerateTrafficData()
			// After cache clear, server load should initially drop
			if loadStr, ok := data["serverLoad"].(string); ok {
				if loadInt, err := strconv.Atoi(strings.TrimSuffix(loadStr, "%")); err == nil {
					reduced := loadInt - (5 + rand.Intn(10))
					if reduced < 10 {
						reduced = 10
					}
					data["serverLoad"] = fmt.Sprintf("%d%%", reduced)
					data["loadPercentage"] = reduced
					data["loadTrend"] = -1 * (5 + rand.Float64()*10)
					data["loadTrendColor"] = colorPositive
					data["loadTrendIcon"] = iconDown
				}
			}

			for key, value := range data {
				dashboard.State.Set(key, value)
			}
		}()

		return nil
	}

	dashboard.Methods["checkSystem"] = func(params map[string]interface{}) error {
		// Simulate system health check
		dashboard.State.Set("wsStatus", "CHECKING")
		dashboard.State.Set("wsStatusColor", colorWarning)
		dashboard.State.Set("wsStatusTextColor", textColorWarning)

		dashboard.State.Set("dbStatus", "CHECKING")
		dashboard.State.Set("dbStatusColor", colorWarning)
		dashboard.State.Set("dbStatusTextColor", textColorWarning)

		dashboard.State.Set("cacheStatus", "CHECKING")
		dashboard.State.Set("cacheStatusColor", colorWarning)
		dashboard.State.Set("cacheStatusTextColor", textColorWarning)

		dashboard.State.Set("notification", "System health check in progress...")

		// Simulate health check with delay
		go func() {
			time.Sleep(2 * time.Second)

			// Randomly generate system status - mostly healthy but occasionally show warnings/errors

			// WebSocket status
			wsRand := rand.Float32()
			if wsRand > 0.15 {
				dashboard.State.Set("wsStatus", "HEALTHY")
				dashboard.State.Set("wsStatusColor", colorHealthy)
				dashboard.State.Set("wsStatusTextColor", textColorHealthy)
			} else if wsRand > 0.05 {
				dashboard.State.Set("wsStatus", "WARNING")
				dashboard.State.Set("wsStatusColor", colorWarning)
				dashboard.State.Set("wsStatusTextColor", textColorWarning)
			} else {
				dashboard.State.Set("wsStatus", "ERROR")
				dashboard.State.Set("wsStatusColor", colorError)
				dashboard.State.Set("wsStatusTextColor", textColorError)
			}

			// Database status
			dbRand := rand.Float32()
			if dbRand > 0.15 {
				dashboard.State.Set("dbStatus", "HEALTHY")
				dashboard.State.Set("dbStatusColor", colorHealthy)
				dashboard.State.Set("dbStatusTextColor", textColorHealthy)
			} else if dbRand > 0.05 {
				dashboard.State.Set("dbStatus", "WARNING")
				dashboard.State.Set("dbStatusColor", colorWarning)
				dashboard.State.Set("dbStatusTextColor", textColorWarning)
			} else {
				dashboard.State.Set("dbStatus", "ERROR")
				dashboard.State.Set("dbStatusColor", colorError)
				dashboard.State.Set("dbStatusTextColor", textColorError)
			}

			// Cache status
			cacheRand := rand.Float32()
			if cacheRand > 0.15 {
				dashboard.State.Set("cacheStatus", "HEALTHY")
				dashboard.State.Set("cacheStatusColor", colorHealthy)
				dashboard.State.Set("cacheStatusTextColor", textColorHealthy)
			} else if cacheRand > 0.05 {
				dashboard.State.Set("cacheStatus", "WARNING")
				dashboard.State.Set("cacheStatusColor", colorWarning)
				dashboard.State.Set("cacheStatusTextColor", textColorWarning)
			} else {
				dashboard.State.Set("cacheStatus", "ERROR")
				dashboard.State.Set("cacheStatusColor", colorError)
				dashboard.State.Set("cacheStatusTextColor", textColorError)
			}

			dashboard.State.Set("lastUpdated", time.Now().Format("Jan 2, 2006 15:04:05"))

			// Generate appropriate notification based on overall system health
			if wsRand > 0.15 && dbRand > 0.15 && cacheRand > 0.15 {
				dashboard.State.Set("notification", "System health check completed: All systems healthy!")
			} else if wsRand > 0.05 && dbRand > 0.05 && cacheRand > 0.05 {
				dashboard.State.Set("notification", "System health check completed: Some warnings detected.")
			} else {
				dashboard.State.Set("notification", "System health check completed: Critical issues detected!")
			}
		}()

		return nil
	}

	return dashboard
}

// GetDashboardStyles returns styles for the admin dashboard
func GetDashboardStyles() string {
	return dashboardStyles
}

// GetDashboardScripts returns scripts for the admin dashboard
func GetDashboardScripts() string {
	return dashboardScript
}
