package server

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			if server == nil {
				t.Fatal("New() returned nil")
			}

			if server.devices == nil {
				t.Error("devices map is nil")
			}

			if len(server.devices) != 0 {
				t.Errorf("devices length = %v, want 0", len(server.devices))
			}

			if server.securityKey != "" {
				t.Errorf("securityKey = %v, want empty", server.securityKey)
			}
		})
	}
}

func TestNewWithSecurityKey(t *testing.T) {
	tests := []struct {
		name        string
		securityKey string
	}{
		{
			name:        "create server with security key",
			securityKey: "secret-key-123",
		},
		{
			name:        "create server with empty security key",
			securityKey: "",
		},
		{
			name:        "create server with long security key",
			securityKey: "very-long-security-key-with-many-characters-for-authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewWithSecurityKey(tt.securityKey)

			if server == nil {
				t.Fatal("NewWithSecurityKey() returned nil")
			}

			if server.devices == nil {
				t.Error("devices map is nil")
			}

			if len(server.devices) != 0 {
				t.Errorf("devices length = %v, want 0", len(server.devices))
			}

			if server.securityKey != tt.securityKey {
				t.Errorf("securityKey = %v, want %v", server.securityKey, tt.securityKey)
			}
		})
	}
}

func TestServer_AddDevice(t *testing.T) {
	tests := []struct {
		name      string
		deviceIDs []string
	}{
		{
			name:      "add single device",
			deviceIDs: []string{"device-001"},
		},
		{
			name:      "add multiple devices",
			deviceIDs: []string{"device-001", "device-002", "device-003"},
		},
		{
			name:      "add device with empty id",
			deviceIDs: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			for _, deviceID := range tt.deviceIDs {
				device := NewDevice()
				device.ID = deviceID
				server.AddDevice(device)
			}

			if len(server.devices) != len(tt.deviceIDs) {
				t.Errorf("devices length = %v, want %v", len(server.devices), len(tt.deviceIDs))
			}

			for _, deviceID := range tt.deviceIDs {
				if _, exists := server.devices[deviceID]; !exists {
					t.Errorf("Device %v not found in server", deviceID)
				}
			}
		})
	}
}

func TestServer_GetDevice(t *testing.T) {
	tests := []struct {
		name        string
		addDevices  []string
		getDeviceID string
		wantExists  bool
	}{
		{
			name:        "get existing device",
			addDevices:  []string{"device-001", "device-002"},
			getDeviceID: "device-001",
			wantExists:  true,
		},
		{
			name:        "get non-existing device",
			addDevices:  []string{"device-001"},
			getDeviceID: "device-999",
			wantExists:  false,
		},
		{
			name:        "get from empty devices",
			addDevices:  []string{},
			getDeviceID: "device-001",
			wantExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			// Add devices
			for _, deviceID := range tt.addDevices {
				device := NewDevice()
				device.ID = deviceID
				server.AddDevice(device)
			}

			// Get device
			device, exists := server.GetDevice(tt.getDeviceID)

			if exists != tt.wantExists {
				t.Errorf("GetDevice() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists && device == nil {
				t.Error("GetDevice() returned nil device when it should exist")
			}

			if tt.wantExists && device.ID != tt.getDeviceID {
				t.Errorf("GetDevice() device.ID = %v, want %v", device.ID, tt.getDeviceID)
			}

			if !tt.wantExists && device != nil {
				t.Error("GetDevice() returned non-nil device when it should not exist")
			}
		})
	}
}

func TestServer_RemoveDevice(t *testing.T) {
	tests := []struct {
		name           string
		addDevices     []string
		removeDeviceID string
		wantRemaining  int
	}{
		{
			name:           "remove existing device",
			addDevices:     []string{"device-001", "device-002", "device-003"},
			removeDeviceID: "device-002",
			wantRemaining:  2,
		},
		{
			name:           "remove non-existing device",
			addDevices:     []string{"device-001"},
			removeDeviceID: "device-999",
			wantRemaining:  1,
		},
		{
			name:           "remove from empty devices",
			addDevices:     []string{},
			removeDeviceID: "device-001",
			wantRemaining:  0,
		},
		{
			name:           "remove last device",
			addDevices:     []string{"device-001"},
			removeDeviceID: "device-001",
			wantRemaining:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			// Add devices
			for _, deviceID := range tt.addDevices {
				device := NewDevice()
				device.ID = deviceID
				server.AddDevice(device)
			}

			// Remove device
			server.RemoveDevice(tt.removeDeviceID)

			// Verify remaining count
			if len(server.devices) != tt.wantRemaining {
				t.Errorf("devices length after remove = %v, want %v", len(server.devices), tt.wantRemaining)
			}

			// Verify removed device is gone
			if _, exists := server.devices[tt.removeDeviceID]; exists {
				t.Errorf("Device %v still exists after removal", tt.removeDeviceID)
			}
		})
	}
}

func TestServer_ListDevices(t *testing.T) {
	tests := []struct {
		name       string
		addDevices []string
		wantCount  int
	}{
		{
			name:       "list with multiple devices",
			addDevices: []string{"device-001", "device-002", "device-003"},
			wantCount:  3,
		},
		{
			name:       "list with single device",
			addDevices: []string{"device-001"},
			wantCount:  1,
		},
		{
			name:       "list with no devices",
			addDevices: []string{},
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			// Add devices
			for _, deviceID := range tt.addDevices {
				device := NewDevice()
				device.ID = deviceID
				server.AddDevice(device)
			}

			// List devices
			ids := server.ListDevices()

			if len(ids) != tt.wantCount {
				t.Errorf("ListDevices() length = %v, want %v", len(ids), tt.wantCount)
			}

			// Verify all expected devices are in the list
			deviceMap := make(map[string]bool)
			for _, id := range ids {
				deviceMap[id] = true
			}

			for _, expectedID := range tt.addDevices {
				if !deviceMap[expectedID] {
					t.Errorf("Device %v not found in ListDevices() result", expectedID)
				}
			}
		})
	}
}

func TestServer_MaxDevicesEviction(t *testing.T) {
	tests := []struct {
		name        string
		numDevices  int
		wantEvicted bool
	}{
		{
			name:        "below max devices",
			numDevices:  10,
			wantEvicted: false,
		},
		{
			name:        "at max devices",
			numDevices:  MaxDevices,
			wantEvicted: false,
		},
		{
			name:        "above max devices triggers eviction",
			numDevices:  MaxDevices + 1,
			wantEvicted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			// Add devices with different connection times
			var firstDeviceID string

			for i := range tt.numDevices {
				device := NewDevice()

				device.ID = string(rune('A' + i))
				if i == 0 {
					firstDeviceID = device.ID
					device.ConnectedAt = time.Now().Add(-time.Hour) // Oldest device
				} else {
					device.ConnectedAt = time.Now().Add(time.Duration(i) * time.Second)
				}

				server.AddDevice(device)
				time.Sleep(1 * time.Millisecond) // Small delay to ensure different timestamps
			}

			// Check if eviction happened
			if tt.wantEvicted {
				// Should have MaxDevices, not more
				if len(server.devices) != MaxDevices {
					t.Errorf("devices length = %v, want %v (after eviction)", len(server.devices), MaxDevices)
				}

				// Oldest device should be evicted
				if _, exists := server.GetDevice(firstDeviceID); exists {
					t.Error("Oldest device was not evicted")
				}
			} else if len(server.devices) != tt.numDevices {
				t.Errorf("devices length = %v, want %v", len(server.devices), tt.numDevices)
			}
		})
	}
}

func TestServer_ConcurrentDeviceAccess(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numOperations int
	}{
		{
			name:          "concurrent add and get",
			numGoroutines: 10,
			numOperations: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()

			done := make(chan bool, tt.numGoroutines)

			// Spawn goroutines that add and get devices concurrently
			for i := range tt.numGoroutines {
				go func(id int) {
					for j := range tt.numOperations {
						deviceID := "device-" + string(rune(id)) + "-" + string(rune(j))
						device := NewDevice()
						device.ID = deviceID

						server.AddDevice(device)
						server.GetDevice(deviceID)
						server.ListDevices()

						if j%2 == 0 {
							server.RemoveDevice(deviceID)
						}
					}

					done <- true
				}(i)
			}

			// Wait for all goroutines to complete
			for range tt.numGoroutines {
				<-done
			}
			// Test passes if no race conditions occurred
		})
	}
}
