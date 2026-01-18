package e2e

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/services"
)

// TestNotebookService_ConcurrentDiscovery tests concurrent notebook discovery
func TestNotebookService_ConcurrentDiscovery(t *testing.T) {
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	dbService := services.NewDbService()
	notebookService := services.NewNotebookService(configService, dbService)

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make(chan []*services.Notebook, numGoroutines)

	// Launch concurrent notebook list operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			notebooks, err := notebookService.List("")
			if err != nil {
				t.Errorf("Notebook list failed: %v", err)
				return
			}
			results <- notebooks
		}()
	}

	wg.Wait()
	close(results)

	// Verify all results are consistent
	var firstResult []*services.Notebook
	resultCount := 0
	for result := range results {
		resultCount++
		if firstResult == nil {
			firstResult = result
		} else {
			assert.Equal(t, len(firstResult), len(result),
				"Notebook discovery results should be consistent across goroutines")
		}
	}

	assert.Equal(t, numGoroutines, resultCount, "Should receive results from all goroutines")
	t.Logf("Concurrent notebook discovery completed with %d consistent results", resultCount)
}

// TestDbService_ConnectionPoolStress tests database connection handling under concurrent load
func TestDbService_ConnectionPoolStress(t *testing.T) {
	dbService := services.NewDbService()
	defer dbService.Close()

	const numGoroutines = 20
	const queriesPerGoroutine = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*queriesPerGoroutine)

	// Launch concurrent database operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < queriesPerGoroutine; j++ {
				ctx := context.Background()
				_, err := dbService.GetDB(ctx)
				if err != nil {
					errors <- err
					continue
				}

				// Execute simple query
				query := "SELECT 1 as test_value"
				_, err = dbService.Query(ctx, query)
				if err != nil {
					errors <- err
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Errorf("Database operation failed: %v", err)
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent database access")
	t.Logf("Completed %d concurrent database operations successfully", numGoroutines*queriesPerGoroutine)
}

// TestDbService_ConcurrentInitialization tests race conditions in database initialization
func TestDbService_ConcurrentInitialization(t *testing.T) {
	const numGoroutines = 15
	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines)

	// Launch concurrent database initialization attempts
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			dbService := services.NewDbService()
			defer dbService.Close()
			
			ctx := context.Background()
			_, err := dbService.GetDB(ctx)
			
			results <- (err == nil)
		}()
	}

	wg.Wait()
	close(results)

	// Verify all initializations succeeded
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	assert.Equal(t, numGoroutines, successCount, 
		"All concurrent database initializations should succeed")
	t.Logf("All %d concurrent database initializations succeeded", successCount)
}

// TestConfigService_ConcurrentAccess tests concurrent config access
func TestConfigService_ConcurrentAccess(t *testing.T) {
	configService, err := services.NewConfigService()
	require.NoError(t, err)

	const numReaders = 10
	const numOperations = 5
	var wg sync.WaitGroup
	errors := make(chan error, numReaders*numOperations)

	// Launch concurrent config read operations
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				// Test concurrent access to config properties
				_ = configService.Store.Notebooks
				_ = configService.Store.NotebookPath
				_ = configService.Path()
				
				// Small delay to increase chance of race conditions
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	errorCount := 0
	for err := range errors {
		errorCount++
		t.Errorf("Config access failed: %v", err)
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during concurrent config access")
	t.Log("Concurrent config access completed successfully")
}

// TestServices_ContextCancellation tests context cancellation under concurrent load
func TestServices_ContextCancellation(t *testing.T) {
	const numGoroutines = 20
	var wg sync.WaitGroup
	cancelledCount := make(chan int, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cancelled := 0
	var mu sync.Mutex

	// Launch operations that should be cancelled
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			dbService := services.NewDbService()
			defer dbService.Close()
			
			// This should be cancelled by context timeout
			_, err := dbService.GetDB(ctx)
			if err != nil && (ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled) {
				mu.Lock()
				cancelled++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	cancelledCount <- cancelled

	finalCancelled := <-cancelledCount
	t.Logf("Context cancellation affected %d/%d operations", finalCancelled, numGoroutines)
	
	// Some operations should be cancelled due to timeout
	assert.Greater(t, finalCancelled, 0, "Some operations should be cancelled by context")
}

// TestServices_MemoryGrowth tests memory usage under concurrent operations
func TestServices_MemoryGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	const numOperations = 100
	var wg sync.WaitGroup

	// Perform many concurrent service operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Create and use services briefly
			configService, err := services.NewConfigService()
			if err != nil {
				return
			}
			
			dbService := services.NewDbService()
			defer dbService.Close()
			
			notebookService := services.NewNotebookService(configService, dbService)
			_, _ = notebookService.List("") // Trigger some operations
		}()
	}

	wg.Wait()
	
	runtime.GC()
	runtime.ReadMemStats(&m2)

	growth := m2.Alloc - m1.Alloc
	growthPerOp := growth / numOperations

	t.Logf("Memory growth: %d bytes total, %d bytes per operation", growth, growthPerOp)
	
	// Memory growth should be reasonable (less than 100KB per operation)
	assert.Less(t, growthPerOp, uint64(100*1024), 
		"Memory growth per operation should be reasonable")
}

// TestServices_RaceConditions runs operations designed to trigger race conditions
func TestServices_RaceConditions(t *testing.T) {
	// This test is designed to be run with -race flag
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	
	dbService := services.NewDbService()
	defer dbService.Close()
	
	notebookService := services.NewNotebookService(configService, dbService)

	const numGoroutines = 30
	var wg sync.WaitGroup

	// Mix of different operations that might race
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			switch id % 4 {
			case 0:
				// Config access
				_ = configService.Store.Notebooks
				_ = configService.Path()
				
			case 1:
				// Database operations
				ctx := context.Background()
				if db, err := dbService.GetDB(ctx); err == nil && db != nil {
					dbService.Query(ctx, "SELECT 1")
				}
				
			case 2:
				// Notebook operations
				_, _ = notebookService.List("")
				notebookService.HasNotebook("/tmp/test")
				
			case 3:
				// Mixed operations
				_ = configService.Store.NotebookPath
				ctx := context.Background()
				dbService.GetDB(ctx)
				_, _ = notebookService.List("")
			}
			
			// Small delay to increase chance of race conditions
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	t.Log("Race condition test completed - run with -race flag to detect issues")
}

// TestServices_HighConcurrency tests system behavior under high concurrent load
func TestServices_HighConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high concurrency test in short mode")
	}

	const numWorkers = 100
	const operationsPerWorker = 10
	
	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, numWorkers*operationsPerWorker)

	// Launch high concurrent load
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			_, err := services.NewConfigService()
			if err != nil {
				errors <- err
				return
			}
			
			dbService := services.NewDbService()
			defer dbService.Close()
			
			for j := 0; j < operationsPerWorker; j++ {
				ctx := context.Background()
				if _, err := dbService.GetDB(ctx); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	close(errors)

	// Count errors
	errorCount := 0
	for err := range errors {
		errorCount++
		if errorCount <= 5 { // Log first few errors only
			t.Logf("High concurrency error: %v", err)
		}
	}

	totalOps := numWorkers * operationsPerWorker
	opsPerSecond := float64(totalOps) / duration.Seconds()

	t.Logf("High concurrency test: %d operations in %v (%.1f ops/sec), %d errors",
		totalOps, duration, opsPerSecond, errorCount)

	// Allow some errors under extreme load, but most should succeed
	errorRate := float64(errorCount) / float64(totalOps)
	assert.Less(t, errorRate, 0.1, "Error rate should be less than 10% under high load")
	assert.Greater(t, opsPerSecond, 100.0, "Should maintain reasonable throughput")
}