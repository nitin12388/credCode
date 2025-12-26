# TrueCaller System - High-Level Design Architecture

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Diagram Overview](#architecture-diagram-overview)
3. [Component Descriptions](#component-descriptions)
4. [Database Schemas](#database-schemas)
5. [Data Flow Explanations](#data-flow-explanations)
6. [Technology Stack](#technology-stack)

---

## 1. System Overview

TrueCaller is a distributed caller identification and spam detection system designed to:
- Identify unknown callers in real-time
- Detect and block spam calls
- Search phone numbers and contacts
- Manage user contacts and profiles
- Provide spam reporting capabilities

The system is built using a microservices architecture with specialized databases for different use cases, connected via message queues for asynchronous processing.

### Key Innovations in This Architecture

1. **Bloom Filter First-Line Defense**: 80-90% of spam checks eliminated by checking if caller is in verified contacts (Bloom Filter), reducing load by 10x.

2. **Spam Score Cache (Identity DB)**: Aerospike-based cache stores pre-computed spam scores with 7-day TTL, providing <2ms lookups instead of expensive ML inference.

3. **Real-time Aggregation (Flink)**: Continuous stream processing of call logs to maintain up-to-date aggregated metrics in Aggregation DB.

4. **Proactive Batch Scoring (Spark)**: Hourly jobs pre-compute spam scores for all numbers from recent calls, keeping cache warm and reducing real-time ML load.

5. **Hybrid ML Inference**: On-demand (for cache misses) + Scheduled batch (for cache warming), optimizing for both latency and cost.

6. **Cost-Optimized Database Choices**: Aerospike for high-performance needs saves ~$70K/month compared to DynamoDB alternatives.

---

## 2. Architecture Diagram Overview

The architecture consists of the following major layers:
- **Client Layer**: User Phone (mobile application)
- **Service Layer**: User Service, Search Service, Identity Service, Call Intercept Service
- **Data Layer**: User DB (PostgreSQL), Search DB (Elasticsearch), Identity DB (Aerospike)
- **Caching Layer**: Redis for performance optimization
- **Messaging Layer**: Kafka for event-driven architecture
- **ML/Analytics Layer**: Feature Store, ML Models Framework, Aggregators
- **Storage Layer**: Dataswarm House for data warehousing

---

## 3. Component Descriptions

### 3.1 User Phone (Mobile Client)

**Purpose**: The mobile application installed on user devices that serves as the primary interface.

**Responsibilities**:
- Capture incoming call events
- Display caller information to users
- Sync user contacts with backend
- Authenticate users
- Allow spam reporting
- Search for phone numbers

**Interactions**:
- Sends API requests to User Service, Search Service, and Call Intercept Service
- Receives real-time caller identification data
- Uploads contact lists for synchronization

---

### 3.2 User Service

**Purpose**: Central service for managing user accounts, profiles, and contact synchronization.

**Responsibilities**:
- **User Registration & Authentication**: Handle user sign-up and login flows
- **Contact Synchronization**: Accept and process contact lists from user phones
- **Profile Management**: Store and update user information
- **Device Management**: Track user devices and sessions

**Key Operations**:
1. **Login**: Authenticate users and create sessions
2. **Contact Sync**: Process uploaded contacts and store in User DB
3. **Profile Updates**: Manage user profile information

**Interactions**:
- Communicates with **User DB** to store user data and contacts
- Communicates with **Identity Service** for authentication
- Publishes events to **Kafka** for contact sync processing
- Queries **Cache** for frequently accessed data

**Data Management**:
- Stores user profiles with phone numbers, device IDs, and metadata
- Maintains contact lists per user with contact phone numbers and names
- Tracks user activity and last sync timestamps

---

### 3.3 Search Service

**Purpose**: Provides fast search capabilities for phone numbers and names.

**Responsibilities**:
- **Phone Number Search**: Look up information for a given phone number
- **Name Search**: Search for phone numbers by contact name
- **Result Aggregation**: Combine data from multiple sources
- **Cache Management**: Implement caching strategy for hot data

**Key Operations**:
1. **Search by Phone**: Primary use case during incoming calls
2. **Search by Name**: User-initiated searches in the app

**Interactions**:
- Queries **Search DB** (Elasticsearch) for indexed search data
- Checks **Cache** (Redis) before hitting database
- Updates cache with search results
- Uses **CDC (Change Data Capture) logs** to keep search index updated

**Performance Characteristics**:
- Sub-100ms response time for p99 requests
- Handles high read throughput
- Implements multi-level caching

---

### 3.4 Identity Service

**Purpose**: Provides spam score lookup for phone numbers during call screening.

**Responsibilities**:
- **Spam Score Lookup**: Fast retrieval of cached spam scores
- **Verified User Check**: Check if number is in verified contacts via Bloom Filter
- **Score Computation**: Trigger full computation when cache miss occurs
- **Cache Management**: Manage spam score cache in Identity DB

**Key Operations**:
1. **Get Spam Score**: Primary operation - check if caller is spam
   - Step 1: Check Bloom Filter (verified users)
   - Step 2: Check Identity DB cache
   - Step 3: Full computation if needed

2. **Compute Spam Score** (on cache miss):
   - Fetch aggregated data from Aggregation DB
   - Call ML Model API with features
   - Store result in Identity DB
   - Return spam score

**Interactions**:
- Queries **Bloom Filter** for verified user check (fastest path)
- Queries **Identity DB** (Aerospike) for cached spam scores
- Reads from **Aggregation DB** (Aerospike) for features
- Calls **ML Models Framework** for inference
- Updates **Identity DB** with computed scores

**Performance Characteristics**:
- Bloom Filter check: <1ms
- Identity DB lookup: <2ms
- Full computation: <50ms (when cache miss)
- Cache hit rate target: 90%+

---

### 3.5 Call Intercept Service

**Purpose**: Real-time processing of incoming calls to identify callers and detect spam.

**Responsibilities**:
- **Real-time Call Processing**: Handle incoming call events
- **Caller Identification**: Look up caller information
- **Spam Detection**: Evaluate spam probability using ML models
- **Spam Reporting**: Process user spam reports

**Key Operations**:
1. **Call Intercept**: Process incoming call and return caller info within milliseconds
2. **Spam Report**: Accept and process spam reports from users

**Interactions**:
- Queries **User DB** and **Search DB** for caller information
- Calls **Identity Service** to validate requests
- Uses **Feature Store** to retrieve ML features
- Invokes **ML Models** for spam prediction
- Publishes events to **Kafka** for analytics and training

**Performance Requirements**:
- Ultra-low latency (<50ms p99)
- High availability (99.99% uptime)
- Real-time processing capability

---

### 3.6 User Database (PostgreSQL)

**Purpose**: Persistent storage for user profiles and contact information.

**Technology Choice**: PostgreSQL

**Rationale**:
- ACID compliance for transactional consistency
- Excellent support for complex relational queries
- Mature ecosystem with proven reliability
- Efficient handling of structured data
- Support for JSON columns for flexible data

**Data Stored**:
- User profiles (phone numbers, device IDs, registration info)
- User contact lists
- User preferences and settings
- Device information

**Scaling Strategy**:
- Vertical scaling for initial growth
- Read replicas for read-heavy workloads
- Partitioning by phone number hash for horizontal scaling
- Connection pooling (PgBouncer)

**Interactions**:
- Primary data source for **User Service**
- Queried by **Call Intercept Service** for user lookups
- Backed up regularly to **Dataswarm House**

---

### 3.7 Search Database (Elasticsearch)

**Purpose**: Optimized search index for fast phone number and name lookups.

**Technology Choice**: Elasticsearch

**Rationale**:
- Inverted index for fast text searches
- Fuzzy matching capabilities
- Horizontal scalability
- Real-time indexing
- Support for complex queries

**Data Stored**:
- Phone numbers as primary keys
- Associated names (from multiple sources)
- Spam scores
- Category information (business/personal)
- Verification status

**Indexing Strategy**:
- Phone number index with tokenization
- Name index with n-gram analysis for fuzzy matching
- Regular index refreshes from User DB via CDC

**Interactions**:
- Primary data source for **Search Service**
- Updated via **CDC logs** from User DB
- Queried during call intercept for caller identification

---

### 3.8 Identity Database (Aerospike) - Spam Score Cache

**Purpose**: Ultra-fast cache for spam scores of phone numbers.

**Technology Choice**: Aerospike

**Rationale**:
- Sub-millisecond latency (p99 < 2ms)
- High throughput (100K+ ops/sec per node)
- Built-in TTL for automatic cache expiration
- Cost-effective compared to DynamoDB (~$3K/month vs ~$40K/month)
- Simple key-value lookups for maximum performance

**Capacity**:
- Billions of phone numbers
- Small record size (~100-200 bytes per record)
- Efficient storage with compression
- Replication Factor: 2 for high availability

**Data Stored**:
- **Key**: Phone number (string)
- **Value**: Spam score object
  - spam_score (float): 0.0 to 1.0
  - last_updated (timestamp)
  - confidence (float)
  - category (string): spam/suspicious/legitimate

**Update Mechanism**:
- **Real-time writes**: After on-demand computation (cache miss)
- **Batch updates**: Spark jobs run every 1 hour
  - Process new call logs from last hour
  - Invoke ML models with features
  - Bulk update spam scores

**Configuration**:
- 5-7 node cluster with i4i.xlarge instances
- SSD storage for low latency
- Hybrid memory mode (index in memory, data on SSD)
- TTL: 7 days (scores older than 7 days are recomputed)

**Interactions**:
- Primary data source for **Identity Service** (spam lookups)
- Updated by **Spark batch jobs** (hourly bulk updates)
- Updated by **Identity Service** (on-demand computations)
- Read-heavy workload (90% reads, 10% writes)

---

### 3.9 Aggregation Database (Aerospike)

**Purpose**: Store real-time aggregated call data and features for ML model inference.

**Technology Choice**: Aerospike

**Rationale**:
- Ultra-low latency for real-time feature serving
- High write throughput for continuous aggregations from Flink
- Efficient storage for time-series aggregated data
- Cost-effective for high-volume operations
- Strong consistency for accurate aggregations

**Data Stored**:
- **Key**: Phone number (string)
- **Value**: Aggregated metrics object
  - call_count_1h (integer): Calls in last 1 hour
  - call_count_24h (integer): Calls in last 24 hours
  - call_count_7d (integer): Calls in last 7 days
  - call_count_30d (integer): Calls in last 30 days
  - unique_callers_7d (integer): Unique numbers called
  - spam_report_count (integer): Total spam reports
  - avg_call_duration (float): Average call duration
  - call_time_pattern (array): Distribution by hour
  - geographic_diversity (float): Call pattern diversity
  - last_call_timestamp (timestamp)
  - features (map): Additional computed features

**Update Mechanism**:
- **Real-time updates**: Flink streaming jobs consume call log events
  - Aggregate metrics in sliding windows
  - Write to Aggregation DB continuously
  - Low-latency streaming aggregations

**Configuration**:
- 5-7 node cluster with i4i.xlarge instances
- SSD storage for low latency
- Hybrid memory mode
- TTL: 30 days for historical aggregations

**Interactions**:
- **Written by**: Flink streaming jobs (real-time aggregation)
- **Read by**: 
  - Identity Service (for on-demand spam computation)
  - Spark batch jobs (for ML model features)
  - ML Models Framework (for inference)
- Write-heavy workload (70% writes, 30% reads)

**Performance**:
- Write latency: <5ms (p99)
- Read latency: <2ms (p99)
- Throughput: 50K+ writes/sec per node

---

### 3.10 Bloom Filter - Verified Users

**Purpose**: Probabilistic data structure to quickly check if a phone number exists in any user's contact list.

**Technology**: In-memory Bloom Filter

**Rationale**:
- **Space Efficient**: Stores billions of phone numbers in a few GB of memory
- **Ultra-Fast Lookup**: O(1) time complexity, <1ms response time
- **Probabilistic**: May have false positives (might say user is verified when not), but NEVER false negatives
- **Perfect for First-Level Filter**: Quickly eliminate verified users from spam check

**How It Works**:
1. When users sync contacts, all contact phone numbers are added to Bloom Filter
2. When checking if number is spam:
   - First check: Is number in Bloom Filter?
   - If YES → Number is likely verified → Return "NOT SPAM" immediately
   - If NO → Number definitely not in contacts → Continue spam check

**Bloom Filter Characteristics**:
- **Size**: ~10 billion phone numbers capacity
- **Memory**: ~5-10 GB for false positive rate of 0.1%
- **Hash Functions**: 7-10 hash functions
- **False Positive Rate**: 0.1% (tunable)
- **False Negative Rate**: 0% (guaranteed)

**Update Mechanism**:
- **Real-time updates**: When contact sync happens, add phone numbers to Bloom Filter
- **Rebuild**: Periodic full rebuild (weekly) from User DB to ensure accuracy
- **Distributed**: Can be sharded by phone number prefix for scalability

**Implementation**:
- In-memory data structure in Identity Service
- Can use Redis Bloom Filter module for distributed setup
- Backed by regular snapshots for recovery

**Trade-offs**:
- **False Positive**: 0.1% of spam numbers might be marked as verified
  - Acceptable trade-off for 99.9% reduction in spam checks
- **No Deletion**: Bloom Filters don't support deletion easily
  - Solution: Periodic rebuild from source of truth

**Impact on System**:
- **Reduces Identity DB load**: 80-90% of queries filtered out
- **Reduces ML model invocations**: Most calls are from known contacts
- **Improves latency**: Sub-millisecond check vs multi-millisecond DB lookup
- **Cost savings**: Fewer expensive operations

---

### 3.11 Cache Layer (Redis)

**Purpose**: In-memory caching for frequently accessed data.

**Technology Choice**: Redis

**Rationale**:
- Extremely fast in-memory operations
- Support for various data structures
- TTL-based expiration
- Pub/sub capabilities
- Distributed caching support

**Cached Data**:
- Hot phone numbers (frequently searched)
- Recent search results
- Active user sessions
- Caller information for popular numbers

**Cache Strategy**:
- Write-through caching for critical data
- LRU (Least Recently Used) eviction policy
- TTL configuration:
  - Hot phone numbers: 1 hour
  - Search results: 30 minutes
  - Session data: 24 hours

**Interactions**:
- Used by all services for performance optimization
- Checked before database queries
- Updated after database writes

---

### 3.12 Message Queue (Kafka)

**Purpose**: Event streaming platform for asynchronous processing and service decoupling.

**Technology Choice**: Apache Kafka

**Rationale**:
- High throughput message processing
- Durability and fault tolerance
- Scalable to millions of messages/sec
- Message replay capability
- Strong ecosystem

**Topics**:
1. **contact-sync-events**: Contact uploads for async processing
2. **call-logs-events**: Detailed call logs from all users (NEW - Primary topic)
   - Event: { caller_phone, receiver_phone, timestamp, duration, call_type, outcome }
   - Consumed by: Flink (real-time aggregation), Datawarehouse (storage)
   - Volume: High (all calls generate events)
3. **call-events**: Call intercept events for analytics
4. **spam-reports**: User spam reports for ML training
5. **search-events**: Search queries for analytics
6. **spam-scores**: Computed spam scores from Spark jobs
   - Event: { phone_number, spam_score, confidence, timestamp, features }
   - Produced by: Spark batch jobs (hourly)
   - Consumed by: Identity DB writer service

**Use Cases**:
- Decouple services for better scalability
- Asynchronous processing of heavy operations
- Event sourcing for audit trails
- Data pipeline for analytics and ML

**Interactions**:
- **Producers**: User Service, Search Service, Call Intercept Service
- **Consumers**: Aggregators, ML pipelines, Analytics services

---

### 3.13 Flink Streaming Jobs

**Purpose**: Real-time stream processing of call logs for continuous aggregation.

**Technology Choice**: Apache Flink

**Rationale**:
- True streaming with low latency (milliseconds)
- Stateful stream processing with exactly-once semantics
- Powerful windowing and aggregation capabilities
- High throughput for processing millions of events/sec
- Fault tolerance with checkpointing

**Streaming Jobs**:

1. **Call Aggregation Job**
   - **Input**: call-logs-events from Kafka
   - **Processing**:
     - Windowed aggregations (1 hour, 24 hours, 7 days, 30 days)
     - Count calls per phone number
     - Calculate unique callers
     - Compute average call duration
     - Detect call patterns (time of day, frequency spikes)
   - **Output**: Write to Aggregation DB (Aerospike) in real-time

2. **Spam Pattern Detection Job**
   - **Input**: call-logs-events + spam-reports from Kafka
   - **Processing**:
     - Detect suspicious patterns (many short calls, robocalls)
     - Track numbers with high spam report rates
     - Identify call center patterns
   - **Output**: Update Aggregation DB with pattern flags

3. **Contact Graph Builder Job**
   - **Input**: contact-sync-events from Kafka
   - **Processing**:
     - Build social graph of contacts
     - Update Bloom Filter with verified numbers
     - Calculate network metrics
   - **Output**: Update Bloom Filter and Aggregation DB

**Processing Windows**:
- **Sliding Windows**: 1 hour, 6 hours, 24 hours
- **Session Windows**: Group related calls from same number
- **Tumbling Windows**: Daily, weekly aggregations

**State Management**:
- **State Backend**: RocksDB for large state
- **Checkpointing**: Every 5 minutes
- **State stored**: Intermediate aggregations, counters, patterns

**Performance**:
- Latency: <5 seconds end-to-end (event to DB update)
- Throughput: 100K+ events/sec per job
- Parallelism: Scalable with task slots

**Deployment**:
- Kubernetes-based deployment
- Auto-scaling based on Kafka lag
- Multi-region for high availability

**Interactions**:
- **Consumes from**: Kafka (call-logs-events, spam-reports, contact-sync-events)
- **Writes to**: Aggregation DB (Aerospike)
- **Updates**: Bloom Filter (for verified contacts)
- **Monitors**: Kafka lag, processing latency

---

### 3.14 Feature Store (Aerospike) - DEPRECATED

**Note**: The Feature Store functionality is now merged with the **Aggregation DB (Section 3.9)**. 

The Aggregation DB serves as both:
1. Real-time aggregated metrics storage (from Flink)
2. ML feature store (for model inference)

This consolidation reduces complexity and cost while maintaining performance.

---

### 3.15 ML Models Framework

**Purpose**: Real-time and batch machine learning inference for spam detection and caller identification.

**Models**:

1. **Spam Detection Model**
   - **Input**: Call patterns, user reports, number characteristics, features from Feature Store
   - **Output**: Spam probability score (0-1)
   - **Algorithm**: Gradient Boosting / Neural Network
   - **Latency**: <10ms inference time

2. **Name Prediction Model**
   - **Input**: Phone number patterns, historical data, contact graph
   - **Output**: Predicted name/category
   - **Use Case**: Identify callers not in user's contacts

3. **Aggregation Model**
   - **Purpose**: Combine multiple data sources with confidence scoring
   - **Input**: Multiple signals from different sources
   - **Output**: Unified caller profile with confidence score

**Training Pipeline**:
- Offline training on historical data from **Dataswarm House**
- Online learning from real-time feedback
- Model versioning and A/B testing
- Regular retraining on new spam patterns

**Inference Modes**:

1. **Real-time Inference** (On-demand)
   - Triggered by Identity Service on cache miss
   - Retrieves features from Aggregation DB
   - Returns spam score within <10ms
   - Low volume, high latency sensitivity

2. **Batch Inference** (Scheduled - Hourly)
   - Triggered by Spark jobs
   - Processes all new call logs from last hour
   - Bulk inference for thousands of numbers
   - High volume, throughput optimized
   - Results published to Kafka → Identity DB

**Interactions**:
- **Real-time**: Called by **Identity Service** for on-demand predictions
- **Batch**: Called by **Spark jobs** for hourly bulk predictions
- Retrieves features from **Aggregation DB** (Aerospike)
- Trained on data from **Dataswarm House**
- Receives feedback via **Kafka** events

---

### 3.16 Spark Batch Jobs

**Purpose**: Hourly batch processing for bulk spam score computation and ML model invocation.

**Technology Choice**: Apache Spark

**Rationale**:
- Excellent for batch processing at scale
- Efficient distributed computing
- Native integration with data sources (Kafka, Aerospike, Datawarehouse)
- Rich ML library (MLlib) for model serving
- Cost-effective for scheduled workloads

**Main Job: Hourly Spam Score Update**

**Schedule**: Every 1 hour

**Steps**:
1. **Read Call Logs**
   - Source: Dataswarm House (data warehouse)
   - Filter: All call-logs-events from last 1 hour
   - Data: { phone_numbers, call patterns, timestamps }

2. **Read ML Features**
   - Source: Aggregation DB (Aerospike)
   - For each unique phone number in call logs
   - Fetch: Aggregated metrics (call counts, patterns, spam reports)

3. **Invoke ML Model**
   - Load pre-trained spam detection model
   - Batch inference for all phone numbers
   - Input: Features from Aggregation DB
   - Output: Spam scores (0.0 to 1.0) with confidence

4. **Publish Results to Kafka**
   - Topic: spam-scores
   - Event: { phone_number, spam_score, confidence, timestamp, model_version }
   - Batched publishing for efficiency

5. **Persist to Identity DB**
   - Consumer service reads from spam-scores topic
   - Bulk writes to Identity DB (Aerospike)
   - Updates spam score cache for future lookups

**Additional Jobs**:

1. **Daily Model Retraining**
   - Read historical data from Dataswarm House
   - Train new version of spam detection model
   - Validate and deploy new model

2. **Weekly Bloom Filter Rebuild**
   - Read all contacts from User DB
   - Rebuild Bloom Filter from scratch
   - Deploy new Bloom Filter

**Performance**:
- Processing Time: ~10-15 minutes for 1 hour of data
- Throughput: Millions of phone numbers per hour
- Parallelism: 100+ Spark executors

**Interactions**:
- **Reads from**:
  - Dataswarm House (call logs)
  - Aggregation DB (ML features)
- **Writes to**:
  - Kafka (spam-scores topic)
- **Uses**:
  - ML Models Framework (batch inference)

---

### 3.17 Aggregators (Deprecated - Replaced by Flink + Spark)

**Note**: The Aggregators concept is now replaced by:
- **Flink** (Section 3.13): Real-time stream processing and aggregation
- **Spark** (Section 3.16): Batch processing and ML inference

This separation provides better performance and clearer separation of concerns.

---

### 3.18 Dataswarm House (Data Warehouse)

**Purpose**: Long-term storage and analytics platform for historical data.

**Technology**: Columnar database (e.g., ClickHouse, Snowflake, or BigQuery)

**Rationale**:
- Efficient storage for large volumes
- Fast analytical queries
- Scalable data warehouse
- Support for complex aggregations

**Data Stored**:
- Historical user activity logs
- Call history and patterns
- Search patterns and trends
- Spam reports archive
- ML training datasets

**Use Cases**:
1. **ML Model Training**: Historical data for training models
2. **Business Analytics**: User behavior analysis, trend identification
3. **Reporting**: Dashboards and reports for stakeholders
4. **Data Science**: Exploratory analysis and experimentation

**Interactions**:
- Receives data from all services via **Kafka**
- Sources data for **ML Models Framework** training
- Feeds **Aggregators** for batch feature computation
- Used by analytics and BI tools

---

### 3.19 CDC (Change Data Capture) Logs

**Purpose**: Track changes in User DB and propagate to Search DB for indexing.

**How It Works**:
- Monitors transaction logs in PostgreSQL
- Captures INSERT, UPDATE, DELETE operations
- Streams changes to Search DB in near real-time

**Benefits**:
- Keep Search DB in sync with User DB
- No need for manual synchronization
- Real-time index updates
- Reliable and fault-tolerant

**Technology Options**:
- Debezium (Kafka Connect)
- PostgreSQL logical replication
- Custom CDC implementation

**Data Flow**:
1. User contacts are added/updated in User DB
2. CDC captures the changes
3. Changes are streamed to Search DB
4. Elasticsearch indices are updated
5. New data becomes searchable immediately

---

## 4. Database Schemas

### 4.1 User Database (PostgreSQL)

#### User Table
```sql
CREATE TABLE users (
    phone_number VARCHAR(20) PRIMARY KEY,
    user_device_id VARCHAR(100),
    device_id VARCHAR(100),
    name VARCHAR(255),
    email VARCHAR(255),
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_sync TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_device_id ON users(device_id);
CREATE INDEX idx_registration_date ON users(registration_date);
```

#### Contact List Table
```sql
CREATE TABLE contact_list (
    id BIGSERIAL PRIMARY KEY,
    user_phone_number VARCHAR(20) REFERENCES users(phone_number),
    contact_phone_number VARCHAR(20),
    contact_name VARCHAR(255),
    added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_phone_number, contact_phone_number)
);

CREATE INDEX idx_user_phone ON contact_list(user_phone_number);
CREATE INDEX idx_contact_phone ON contact_list(contact_phone_number);
CREATE INDEX idx_contact_name ON contact_list(contact_name);
```

#### Device Table
```sql
CREATE TABLE devices (
    device_id VARCHAR(100) PRIMARY KEY,
    phone_number VARCHAR(20) REFERENCES users(phone_number),
    device_model VARCHAR(100),
    os_version VARCHAR(50),
    app_version VARCHAR(50),
    last_active TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_phone_number ON devices(phone_number);
```

---

### 4.2 Search Database (Elasticsearch)

#### Phone Number Index
```json
{
  "mappings": {
    "properties": {
      "phone_number": {
        "type": "keyword"
      },
      "names": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword"
          },
          "ngram": {
            "type": "text",
            "analyzer": "ngram_analyzer"
          }
        }
      },
      "spam_score": {
        "type": "float"
      },
      "category": {
        "type": "keyword"
      },
      "verification_status": {
        "type": "keyword"
      },
      "report_count": {
        "type": "integer"
      },
      "last_updated": {
        "type": "date"
      }
    }
  },
  "settings": {
    "analysis": {
      "analyzer": {
        "ngram_analyzer": {
          "type": "custom",
          "tokenizer": "ngram_tokenizer",
          "filter": ["lowercase"]
        }
      },
      "tokenizer": {
        "ngram_tokenizer": {
          "type": "ngram",
          "min_gram": 2,
          "max_gram": 10
        }
      }
    }
  }
}
```

---

### 4.3 Identity Database (Aerospike) - Spam Score Cache

#### Spam Scores Namespace
```
Namespace: spam_scores
Set: phone_scores

Key Structure: {phoneNumber}  (e.g., "1234567890")

Record Structure:
{
  "phone_number": "string",
  "spam_score": "float",          // 0.0 to 1.0
  "confidence": "float",          // 0.0 to 1.0
  "category": "string",           // "spam", "suspicious", "legitimate"
  "last_updated": "timestamp",
  "model_version": "string",
  "source": "string"              // "batch" or "realtime"
}

Bins (Columns):
- phone_number (string) - redundant but useful for debugging
- spam_score (double) - primary value
- confidence (double) - confidence in prediction
- category (string) - classification
- last_updated (integer, timestamp)
- model_version (string) - which model version computed this
- source (string) - batch (Spark) or realtime (on-demand)

TTL: 604800 seconds (7 days)
```

#### Configuration
```
namespace spam_scores {
    replication-factor 2
    memory-size 16G
    storage-engine device {
        device /dev/nvme0n1
        write-block-size 128K
    }
    default-ttl 604800  # 7 days
    
    # Optimize for read-heavy workload
    enable-benchmarks-storage true
}
```

#### Example Operations

**Write (after ML inference)**:
```
Key: "1234567890"
Values: {
  spam_score: 0.85,
  confidence: 0.92,
  category: "spam",
  last_updated: 1703592000,
  model_version: "v2.3",
  source: "batch"
}
```

**Read (during call intercept)**:
```
GET Key: "1234567890"
Returns: { spam_score: 0.85, confidence: 0.92, category: "spam", ... }
Latency: <2ms (p99)
```

---

### 4.4 Aggregation Database / Feature Store (Aerospike)

**Purpose**: Stores real-time aggregated call metrics and ML features, updated continuously by Flink.

#### Aggregated Features Namespace
```
Namespace: aggregations
Set: phone_features

Key Structure: {phoneNumber}  (e.g., "1234567890")

Record Structure:
{
  "phone_number": "string",
  
  // Call frequency metrics (updated by Flink)
  "call_count_1h": "integer",       // Calls in last 1 hour
  "call_count_24h": "integer",      // Calls in last 24 hours
  "call_count_7d": "integer",       // Calls in last 7 days
  "call_count_30d": "integer",      // Calls in last 30 days
  
  // Caller diversity metrics
  "unique_callers_7d": "integer",   // Unique numbers called in 7d
  "unique_callers_30d": "integer",  // Unique numbers called in 30d
  
  // Spam-related metrics
  "spam_report_count": "integer",   // Total spam reports received
  "spam_report_7d": "integer",      // Spam reports in last 7 days
  
  // Call duration metrics
  "avg_call_duration": "float",     // Average call duration in seconds
  "total_call_duration_7d": "float", // Total duration in 7 days
  
  // Temporal patterns (array of 24 integers, one per hour)
  "call_time_pattern": "list",      // Call distribution by hour of day
  
  // Geographic diversity
  "geographic_diversity": "float",  // 0-1 score of geographic spread
  
  // Metadata
  "last_call_timestamp": "timestamp",
  "first_seen_timestamp": "timestamp",
  "last_updated": "timestamp"
}

Bins (Columns):
- phone_number (string)
- call_count_1h (integer)
- call_count_24h (integer)
- call_count_7d (integer)
- call_count_30d (integer)
- unique_callers_7d (integer)
- unique_callers_30d (integer)
- spam_report_count (integer)
- spam_report_7d (integer)
- avg_call_duration (double)
- total_call_duration_7d (double)
- call_time_pattern (list of integers)
- geographic_diversity (double)
- last_call_timestamp (integer, timestamp)
- first_seen_timestamp (integer, timestamp)
- last_updated (integer, timestamp)

TTL: 2592000 seconds (30 days)
```

#### Configuration
```
namespace aggregations {
    replication-factor 2
    memory-size 64G
    storage-engine device {
        device /dev/nvme0n1
        write-block-size 128K
    }
    default-ttl 2592000  # 30 days
    
    # Optimized for write-heavy workload (Flink writes)
    write-block-size 128K
    max-write-cache 256M
}
```

#### Example Operations

**Write (Flink update after processing call event)**:
```
Key: "1234567890"
UPDATE: {
  call_count_1h: INCREMENT 1,
  call_count_24h: INCREMENT 1,
  call_count_7d: INCREMENT 1,
  call_count_30d: INCREMENT 1,
  last_call_timestamp: 1703592000,
  last_updated: 1703592000
}
```

**Read (during ML inference)**:
```
GET Key: "1234567890"
Returns: {
  call_count_1h: 5,
  call_count_24h: 47,
  call_count_7d: 234,
  call_count_30d: 1150,
  spam_report_count: 23,
  avg_call_duration: 45.3,
  ...
}
Latency: <2ms (p99)
```

**Usage in ML Model**:
- These features are input to spam detection model
- High call frequency + high spam reports = likely spam
- Short avg_call_duration + many calls = robocall pattern
- Unusual time patterns = suspicious behavior

---

### 4.5 Bloom Filter (Redis or In-Memory)

**Purpose**: Probabilistic data structure to quickly determine if a phone number exists in any user's contact list.

#### Implementation Option 1: Redis Bloom Filter Module

```
Module: RedisBloom
Command: BF.ADD / BF.EXISTS

Structure:
Key: verified_contacts_bloom
Type: Bloom Filter
```

**Configuration**:
```
BF.RESERVE verified_contacts_bloom 0.001 10000000000
# Error rate: 0.1%
# Expected insertions: 10 billion phone numbers
# Estimated memory: ~5-10 GB
```

**Operations**:

**Add phone number (during contact sync)**:
```
BF.ADD verified_contacts_bloom "1234567890"
Returns: 1 (added) or 0 (already exists)
```

**Check if phone number exists**:
```
BF.EXISTS verified_contacts_bloom "1234567890"
Returns: 1 (might exist) or 0 (definitely doesn't exist)
```

#### Implementation Option 2: In-Memory Bloom Filter

**Language**: Go (using github.com/bits-and-blooms/bloom)

**Configuration**:
```go
// Create Bloom Filter
bloomFilter := bloom.NewWithEstimates(
    10_000_000_000,  // 10 billion expected items
    0.001,           // 0.1% false positive rate
)

// Memory: ~5-8 GB
// Hash functions: 10
```

**Operations**:
```go
// Add phone number
bloomFilter.AddString("1234567890")

// Check existence
exists := bloomFilter.TestString("1234567890")
// returns true (might exist) or false (definitely doesn't)
```

#### Bloom Filter Characteristics

| Parameter | Value |
|-----------|-------|
| Capacity | 10 billion phone numbers |
| False Positive Rate | 0.1% (1 in 1000) |
| False Negative Rate | 0% (guaranteed) |
| Memory Usage | ~5-10 GB |
| Lookup Time | O(1), <1ms |
| Hash Functions | 10 |

#### Update Strategy

**1. Real-time Updates (Incremental)**
- When user syncs contacts, add phone numbers to Bloom Filter
- Continuous updates as users sync
- Eventual consistency is acceptable

**2. Weekly Full Rebuild**
- Every week, rebuild Bloom Filter from scratch
- Read all contacts from User DB
- Build new Bloom Filter
- Atomic swap to new filter
- Ensures accuracy and removes stale entries

**Code Example (Go)**:
```go
// Weekly rebuild job
func rebuildBloomFilter() *bloom.BloomFilter {
    newBloom := bloom.NewWithEstimates(10_000_000_000, 0.001)
    
    // Read all contacts from User DB
    contacts := fetchAllContacts()
    
    for _, phoneNumber := range contacts {
        newBloom.AddString(phoneNumber)
    }
    
    return newBloom
}

// Atomic swap
atomic.StorePointer(&globalBloomFilter, unsafe.Pointer(newBloom))
```

#### False Positive Handling

- **Impact**: 0.1% of non-contacts might be marked as "verified"
- **Result**: Small percentage of spam numbers avoid spam check
- **Mitigation**: Acceptable trade-off for massive performance gain
- **Alternative**: Lower false positive rate (increases memory usage)

---

### 4.6 Cache (Redis)

#### Data Structures

**Hot Phone Numbers Cache**
```
Key: caller:{phoneNumber}
Type: Hash
TTL: 3600 seconds (1 hour)

Fields:
- name: "string"
- spam_score: "float"
- category: "string"
- report_count: "integer"
- last_updated: "timestamp"
```

**Search Results Cache**
```
Key: search:{phoneNumber}
Type: String (JSON)
TTL: 1800 seconds (30 minutes)

Value: JSON string with search results
{
  "phoneNumber": "string",
  "names": ["name1", "name2"],
  "spamScore": "float",
  "category": "string"
}
```

**Session Cache**
```
Key: session:{sessionToken}
Type: Hash
TTL: 86400 seconds (24 hours)

Fields:
- userId: "string"
- phoneNumber: "string"
- deviceId: "string"
- expiresAt: "timestamp"
```

---

## 5. Data Flow Explanations

### 5.1 User Registration and Login Flow

**Step-by-step Process**:

1. **User Initiates Login**
   - User opens app and enters phone number
   - App sends login request to User Service

2. **User Service Processing**
   - User Service validates phone number format
   - Checks if user exists in User DB
   - If new user, creates user record
   - Requests identity verification from Identity Service

3. **Identity Verification**
   - Identity Service generates OTP (One-Time Password)
   - OTP sent to user's phone
   - User enters OTP in app
   - Identity Service validates OTP

4. **Session Creation**
   - Identity Service generates session token and refresh token
   - Tokens stored in Identity DB (Aerospike) with TTL
   - Session also cached in Redis for faster access
   - Tokens returned to User Service, then to app

5. **User Authenticated**
   - App stores tokens securely
   - All subsequent requests include session token
   - User is logged in and can use app features

**Data Flow Diagram**:
```
User Phone → User Service → User DB (create/validate user)
                ↓
         Identity Service → Identity DB (create session)
                ↓
              Cache (cache session)
                ↓
         Return tokens to User Phone
```

---

### 5.2 Contact Synchronization Flow

**Step-by-step Process**:

1. **User Grants Permission**
   - User grants contact access permission to app
   - App reads contacts from phone

2. **Upload Contacts**
   - App sends contact list to User Service
   - Request includes: user's phone number, list of contacts (name + phone)
   - Large contact lists may be batched

3. **User Service Validation**
   - User Service validates session token with Identity Service
   - Validates contact data format
   - Sanitizes and normalizes phone numbers

4. **Store in User DB**
   - Contact list stored in User DB
   - Associated with user's phone number
   - Updates timestamp of last sync

5. **Publish to Kafka**
   - User Service publishes contact-sync event to Kafka
   - Event includes: user phone, contact data, timestamp

6. **Asynchronous Aggregation**
   - Aggregators consume contact-sync events from Kafka
   - Process contacts to extract name-phone mappings
   - Update spam scores based on how many users have the number

7. **Update Search DB**
   - Aggregator updates Search DB (Elasticsearch) with new data
   - Adds/updates phone numbers with associated names
   - Updates indices for fast searching

8. **CDC Updates**
   - CDC logs capture changes in User DB
   - Stream changes to Search DB for real-time indexing
   - Ensures Search DB stays in sync

**Data Flow Diagram**:
```
User Phone → User Service → User DB (store contacts)
                ↓
             Kafka (publish event)
                ↓
          Aggregators (process)
                ↓
         Search DB (update indices)

Parallel:
User DB → CDC Logs → Search DB (real-time sync)
```

---

### 5.3 Phone Number Search Flow

**Step-by-step Process**:

1. **User Searches**
   - User enters phone number in search bar
   - App sends search request to Search Service

2. **Check Cache First**
   - Search Service checks Redis cache
   - Key: `search:{phoneNumber}`
   - If cache hit: return cached results immediately (fastest path)

3. **Cache Miss - Query Search DB**
   - If not in cache, Search Service queries Search DB (Elasticsearch)
   - Elasticsearch performs fast lookup using inverted index
   - Returns matching results with names, spam score, category

4. **Aggregate Results**
   - Search Service may combine data from multiple sources
   - Applies business logic and filtering
   - Formats response

5. **Update Cache**
   - Search Service stores results in Redis cache
   - Sets appropriate TTL (30 minutes)
   - Future requests will hit cache

6. **Log to Kafka**
   - Search event logged to Kafka (search-events topic)
   - Used for analytics and ML training

7. **Return to User**
   - Search results displayed to user
   - Shows caller name, spam score, category

**Data Flow Diagram**:
```
User Phone → Search Service → Cache (check)
                              ↓ (miss)
                          Search DB → return results
                              ↓
                          Cache (update)
                              ↓
                          Kafka (log event)
                              ↓
                        User Phone (display)
```

---

### 5.4 Call Intercept Flow (Most Critical) - NEW ARCHITECTURE

**Step-by-step Process**:

1. **Incoming Call Detected**
   - Phone OS detects incoming call
   - App receives call event with caller's phone number
   - App immediately sends request to Call Intercept Service

2. **Step 1: Bloom Filter Check (Fastest Path)**
   - Identity Service checks Bloom Filter
   - Question: "Is this caller in any user's contact list?"
   - **If YES** (caller is verified):
     - Return immediately: "NOT SPAM" (trusted contact)
     - Skip all remaining checks
     - Latency: <1ms
     - **~80-90% of calls take this path**
   - **If NO** (caller not in contacts):
     - Continue to Step 3

3. **Step 2: Identity DB Check (Fast Path)**
   - Identity Service queries Identity DB (Aerospike)
   - Lookup: spam score for phone number
   - **If FOUND** (cache hit):
     - Return cached spam score
     - Latency: <2ms
     - **~8-10% of calls take this path**
   - **If NOT FOUND** (cache miss):
     - Continue to Step 4

4. **Step 3: Full Spam Computation (Slow Path)**
   - **Only ~1-2% of calls reach here**
   
   a. **Fetch Aggregated Data**
      - Query Aggregation DB (Aerospike)
      - Get features: call counts, patterns, spam reports
      
   b. **Invoke ML Model**
      - Identity Service calls ML Models Framework
      - Real-time inference with fetched features
      - Model returns spam score (0.0 to 1.0)
      
   c. **Store in Identity DB**
      - Cache the computed spam score
      - Key: phone_number
      - Value: { spam_score, confidence, timestamp }
      - TTL: 7 days
      
   d. **Return Result**
      - Return spam score to Call Intercept Service

5. **Get Caller Name** (Parallel to spam check)
   - Call Intercept Service queries Search DB
   - Get caller name if available
   - Can happen in parallel with spam check

6. **Aggregate Response**
   - Combine spam score + caller name
   - Format for display:
     - Caller name (if known)
     - Spam warning (if score > threshold)
     - Category (spam/suspicious/legitimate)

7. **Update Cache (Optional)**
   - Store complete caller info in Redis cache
   - For even faster lookups next time

8. **Log to Kafka**
   - Publish call-logs-event to Kafka
   - Event: { caller_phone, receiver_phone, timestamp, spam_score, outcome }
   - Consumed by Flink for real-time aggregation
   - Stored in Dataswarm House for batch processing

9. **Display to User**
   - Caller info displayed on phone screen during ringing
   - User sees: name, spam warning (if spam), category
   - User can choose to answer or reject

**Performance Characteristics**:

| Path | % of Calls | Latency | Description |
|------|------------|---------|-------------|
| Bloom Filter Hit | 80-90% | <1ms | Verified contacts |
| Identity DB Hit | 8-10% | <2ms | Cached spam scores |
| Full Computation | 1-2% | <50ms | New/unknown numbers |

**Overall Performance**:
- **Average latency**: <5ms (due to Bloom Filter optimization)
- **P99 latency**: <50ms
- **Target SLA**: <50ms for all calls

**Data Flow Diagram**:
```
Incoming Call → User Phone → Call Intercept Service
                                      ↓
                              Identity Service
                                      ↓
                    ┌─────── Bloom Filter (Step 1) ───────┐
                    │                                      │
                [IN FILTER]                           [NOT IN]
                    │                                      │
                Return "NOT SPAM"              Identity DB (Step 2)
                    ↓                                      │
                                                     ┌─────┴─────┐
                                                  [FOUND]    [NOT FOUND]
                                                     │            │
                                              Return Score    Full Compute
                                                     │            │
                                                     │      Aggregation DB
                                                     │            │
                                                     │      ML Models
                                                     │            │
                                                     │      Identity DB (write)
                                                     │            │
                                                     └─────┬──────┘
                                                           ↓
                              Kafka (log call-logs-event)
                                      ↓
                              Display to User Phone
```

**Key Optimization - Bloom Filter**:
- Eliminates 80-90% of lookups (trusted contacts)
- Reduces load on Identity DB by 10x
- Reduces ML model invocations by 100x
- Massive cost savings and latency improvement

---

### 5.5 Spam Reporting Flow

**Step-by-step Process**:

1. **User Reports Spam**
   - After receiving a call, user marks it as spam
   - App sends spam report to Call Intercept Service
   - Includes: reported phone number, reporter's phone, timestamp, reason

2. **Validate Report**
   - Call Intercept Service validates session token
   - Checks if user has already reported this number (prevent duplicates)
   - Validates phone number format

3. **Store in User DB**
   - Spam report stored in User DB for audit trail
   - Updates spam report count for the reported number

4. **Update Search DB**
   - Increment spam report count in Search DB
   - Update spam score (simple aggregation)
   - Update category if threshold reached

5. **Publish to Kafka**
   - Spam report event published to Kafka (spam-reports topic)
   - Event includes full context for ML training

6. **Asynchronous Processing**
   - Aggregators consume spam reports
   - Calculate aggregated spam scores
   - Detect spam patterns and trends

7. **Update Feature Store**
   - Aggregators update Feature Store with new spam metrics
   - Updates features like: spam_report_count, spam_score
   - Used for future ML predictions

8. **ML Model Retraining**
   - Spam reports streamed to Dataswarm House
   - Batch jobs periodically retrain ML models
   - New models deployed to production

9. **Confirmation to User**
   - User receives confirmation that report was submitted
   - May show community impact (e.g., "1000 users have reported this")

**Data Flow Diagram**:
```
User Phone → Call Intercept Service → User DB (store report)
                                     ↓
                                 Search DB (update score)
                                     ↓
                                 Kafka (publish event)
                                     ↓
                              Aggregators (process)
                                     ↓
                         Feature Store (update features)
                                     ↓
                         Dataswarm House (long-term storage)
                                     ↓
                         ML Models (retraining)
```

---

### 5.6 Hourly Spam Score Batch Update Flow (NEW)

**Schedule**: Every 1 hour

**Purpose**: Proactively compute spam scores for all numbers that received calls in the last hour, keeping Identity DB cache warm.

**Step-by-step Process**:

1. **Spark Job Triggered**
   - Scheduled job starts every hour (e.g., at :00 minutes)
   - Job runs on Spark cluster

2. **Read Call Logs from Datawarehouse**
   - Source: Dataswarm House
   - Query: All call-logs-events from last 1 hour
   - Extract: Unique phone numbers (both callers and receivers)
   - Result: List of ~1M phone numbers

3. **Read ML Features from Aggregation DB**
   - For each phone number in the list
   - Bulk read from Aggregation DB (Aerospike)
   - Fetch features:
     - call_count_1h, call_count_24h, call_count_7d, call_count_30d
     - spam_report_count
     - avg_call_duration
     - call_time_pattern
     - geographic_diversity
     - And other aggregated metrics

4. **Batch ML Inference**
   - Load spam detection ML model
   - Batch inference for all phone numbers
   - Input: Feature vectors from Aggregation DB
   - Output: Spam scores (0.0 to 1.0) + confidence
   - Optimized for throughput (not latency)
   - Process millions of numbers in minutes

5. **Publish Results to Kafka**
   - Topic: spam-scores
   - Batch publish events:
     ```json
     {
       "phone_number": "1234567890",
       "spam_score": 0.85,
       "confidence": 0.92,
       "timestamp": "2025-12-26T10:00:00Z",
       "model_version": "v2.3",
       "features_used": ["call_freq", "reports", "patterns"]
     }
     ```
   - High throughput batch publishing

6. **Consumer Service Reads from Kafka**
   - Dedicated consumer service
   - Consumes spam-scores topic
   - Batches events for bulk writes

7. **Bulk Write to Identity DB**
   - Batch writes to Identity DB (Aerospike)
   - Update spam scores for all processed numbers
   - Overwrites existing scores (keeping cache fresh)
   - High throughput bulk operations

8. **Cache Warming Complete**
   - Identity DB now has fresh spam scores
   - Future call intercepts will hit cache (fast path)
   - Reduces real-time ML invocations

**Performance Metrics**:
- **Input**: ~1-2 million phone numbers per hour
- **Processing Time**: 10-15 minutes
- **ML Inference**: ~100K numbers/minute
- **Write Throughput**: ~50K writes/sec to Identity DB

**Benefits**:
1. **Proactive**: Scores computed before next call
2. **Cache Warming**: Keeps Identity DB hot
3. **Cost Effective**: Batch inference is cheaper than real-time
4. **Scalable**: Handles millions of numbers

**Data Flow Diagram**:
```
Hourly Trigger → Spark Job
                     ↓
              Dataswarm House (read call logs)
                     ↓
              Extract unique phone numbers
                     ↓
              Aggregation DB (read features)
                     ↓
              ML Models Framework (batch inference)
                     ↓
              Kafka (publish spam-scores)
                     ↓
              Consumer Service
                     ↓
              Identity DB (bulk write)
                     ↓
              Cache warmed for next hour
```

---

### 5.7 Real-time Call Aggregation Flow (Flink)

**Purpose**: Continuously aggregate call data in real-time for ML feature computation.

**Step-by-step Process**:

1. **Call Log Event Published**
   - Every call generates an event
   - Published to Kafka: call-logs-events topic
   - Event: { caller_phone, receiver_phone, timestamp, duration, outcome }

2. **Flink Consumes Event**
   - Flink streaming job consumes from Kafka
   - Low latency (<5 seconds)
   - Exactly-once processing semantics

3. **Windowed Aggregations**
   - **1-hour window**: Count calls in last 1 hour
   - **24-hour window**: Count calls in last 24 hours
   - **7-day window**: Count calls in last 7 days
   - **30-day window**: Count calls in last 30 days
   - Sliding windows for continuous updates

4. **Compute Additional Metrics**
   - Unique callers count
   - Average call duration
   - Call time patterns (histogram by hour)
   - Geographic diversity (if location data available)
   - Call frequency trends

5. **Write to Aggregation DB**
   - Continuous writes to Aggregation DB (Aerospike)
   - Update existing records or create new
   - Each write updates the aggregated metrics
   - Low latency writes (<5ms)

6. **Features Ready for Consumption**
   - Aggregation DB now has fresh features
   - Available for:
     - Real-time spam computation (Identity Service)
     - Batch processing (Spark jobs)
     - ML model inference

7. **Parallel: Flow to Datawarehouse**
   - Events also flow to Dataswarm House
   - Long-term storage for batch processing
   - Used by Spark jobs for hourly updates

**Performance**:
- **Latency**: <5 seconds (event to aggregation)
- **Throughput**: 100K+ events/sec
- **State Size**: TB-scale (windowed state)

**Data Flow Diagram**:
```
Call Happens → Phone App → Call Intercept Service
                                    ↓
                            Kafka (call-logs-events)
                            ↓                    ↓
                        Flink Jobs       Dataswarm House
                            ↓
                    Windowed Aggregations
                    (1h, 24h, 7d, 30d)
                            ↓
                    Aggregation DB (write)
                            ↓
            Features ready for ML inference
```

---

### 5.8 ML Feature Computation Flow (UPDATED)

**Step-by-step Process**:

1. **Event Collection**
   - All user activities generate events: calls, searches, spam reports, contacts
   - Events published to Kafka topics (especially call-logs-events)

2. **Real-time Stream Processing (Flink)**
   - Flink consumes events in real-time from Kafka
   - Calculate streaming features with sliding windows:
     - Call frequency (1h, 24h, 7d, 30d windows)
     - Unique caller counts
     - Average call duration
     - Call time patterns
   - Update Aggregation DB (Aerospike) continuously
   - Latency: <5 seconds from event to feature

3. **Aggregation DB - Feature Storage**
   - Aggregation DB (Aerospike) serves as the Feature Store
   - Stores real-time aggregated features
   - Available for both real-time and batch ML inference
   - Ultra-low latency reads (<2ms)

4. **Batch Feature Computation (Optional)**
   - Daily Spark jobs for complex features:
     - Network effects (social graph analysis)
     - Long-term trends (beyond 30 days)
     - Cross-number patterns
   - These are supplementary to real-time features

5. **Model Training**
   - Training pipelines read from:
     - Historical data: Dataswarm House
     - Features: Aggregation DB snapshots
   - Train models offline on large datasets
   - Regular retraining (weekly) on new patterns

6. **Model Deployment**
   - Trained models deployed to ML Models Framework
   - Two inference modes:
     - Real-time: On-demand for cache misses
     - Batch: Hourly Spark jobs for bulk scoring
   - Features retrieved from Aggregation DB

**Data Flow Diagram**:
```
Call Events → Kafka (call-logs-events)
                ↓              ↓
            Flink Jobs    Dataswarm House
                ↓              ↓
      Aggregation DB     Historical Storage
      (Feature Store)          ↓
                ↓          Model Training
        ML Inference           ↓
                         Model Deployment
```

**Key Changes from Old Architecture**:
- Flink replaced generic Aggregators for real-time processing
- Aggregation DB serves as Feature Store (consolidated)
- Spark handles batch ML inference (hourly updates)
- Clearer separation: Flink for features, Spark for ML

---

### 5.7 Search Index Update Flow (via CDC)

**Step-by-step Process**:

1. **Data Change in User DB**
   - User adds/updates contacts in User DB
   - Database transaction commits
   - PostgreSQL writes to Write-Ahead Log (WAL)

2. **CDC Captures Change**
   - CDC tool (e.g., Debezium) monitors PostgreSQL WAL
   - Detects INSERT/UPDATE/DELETE operations
   - Captures change data in real-time

3. **Transform Change Event**
   - CDC extracts relevant fields
   - Formats change as event
   - Adds metadata (timestamp, operation type)

4. **Stream to Search DB**
   - Change event streamed to Elasticsearch
   - May go through Kafka for reliability
   - Or directly to Elasticsearch bulk API

5. **Update Search Index**
   - Elasticsearch updates index
   - Re-indexes affected documents
   - Updates inverted index for fast search

6. **Index Refresh**
   - Elasticsearch refreshes index (near real-time)
   - New data becomes searchable within seconds

7. **Verification**
   - CDC monitors for failures
   - Retries failed updates
   - Maintains consistency between User DB and Search DB

**Benefits**:
- Near real-time synchronization
- No manual sync needed
- Reliable and fault-tolerant
- Decoupled systems

**Data Flow Diagram**:
```
User DB (write) → WAL (Write-Ahead Log)
                        ↓
                  CDC Tool (capture)
                        ↓
                  Kafka (optional)
                        ↓
                  Search DB (update index)
                        ↓
                  Index Refresh (searchable)
```

---

## 6. Technology Stack

### 6.1 Programming Languages
- **Go**: Primary language for all services
  - High performance and concurrency
  - Excellent for microservices
  - Strong standard library
  - Fast compilation and deployment

### 6.2 Databases

| Database | Technology | Use Case | Key Features |
|----------|-----------|----------|--------------|
| User DB | PostgreSQL | User profiles, contacts | ACID, relational, mature |
| Search DB | Elasticsearch | Phone/name search | Full-text search, fast lookups |
| Identity DB | Aerospike | Spam score cache | Ultra-low latency, simple key-value |
| Aggregation DB | Aerospike | Real-time aggregated metrics & ML features | High write throughput, low read latency |
| Cache | Redis | Hot data & Bloom Filter | In-memory, fast, TTL support |

### 6.3 Message Queue
- **Apache Kafka**
  - Event streaming
  - High throughput
  - Durable messaging
  - Replay capability

### 6.4 ML/Analytics
- **ML Framework**: TensorFlow / PyTorch
- **Stream Processing**: Apache Flink (real-time aggregations)
- **Batch Processing**: Apache Spark (hourly ML inference, model training)
- **Data Warehouse**: ClickHouse / Snowflake / BigQuery (Dataswarm House)

### 6.5 Infrastructure
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Service Mesh**: Istio / Linkerd
- **API Gateway**: Kong / Envoy

### 6.6 Monitoring & Observability
- **Metrics**: Prometheus
- **Visualization**: Grafana
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Tracing**: Jaeger / Zipkin
- **Alerting**: PagerDuty / Opsgenie

### 6.7 CDC (Change Data Capture)
- **Debezium** (Kafka Connect)
- PostgreSQL Logical Replication
- Custom CDC implementation

---

## 7. Key Design Principles

### 7.1 Performance
- Sub-50ms latency for call intercept (critical path)
- Sub-100ms for search operations
- Multi-level caching strategy
- Parallel database queries
- Pre-computed ML features

### 7.2 Scalability
- Microservices architecture for independent scaling
- Stateless services (can scale horizontally)
- Database sharding for data growth
- Kafka for decoupling and async processing

### 7.3 Reliability
- High availability (99.99% target)
- Replication for critical databases
- Circuit breakers for fault tolerance
- Graceful degradation

### 7.4 Cost Optimization
- Aerospike chosen over DynamoDB (saves ~$70K/month)
- Efficient caching reduces database load
- Batch processing for non-real-time operations
- Resource optimization with Kubernetes

### 7.5 Security & Privacy
- JWT-based authentication
- Data encryption at rest and in transit
- Minimal PII storage
- GDPR compliance
- User consent for data usage

---

## 8. System Characteristics

### 8.1 Performance SLAs
- **Call Intercept**: <50ms (p99)
- **Search API**: <100ms (p99)
- **Contact Sync**: <5s for 1000 contacts
- **Uptime**: 99.99%

### 8.2 Capacity Planning
- **Users**: Designed for 100M+ users
- **Daily Active Users**: 10M+
- **Calls Processed**: 100M+ per day
- **Searches**: 50M+ per day
- **Database Size**: Multi-TB scale

### 8.3 Traffic Patterns
- **Peak Hours**: Morning and evening commute times
- **Geographic Distribution**: Global, follow-the-sun pattern
- **Spikes**: During spam campaigns, special events

---

## Summary

This HLD architecture for TrueCaller is designed for:
- **Ultra-low latency** for real-time call identification (<5ms average, <50ms p99)
- **High scalability** to handle millions of users and billions of phone numbers
- **Cost efficiency** through smart technology choices (Aerospike over DynamoDB saves $70K/month)
- **ML-powered** spam detection with real-time and batch inference
- **Fault tolerance** and high availability
- **Data privacy** and security compliance

## Key Architectural Highlights

### 1. Three-Tier Spam Check Strategy
- **Tier 1**: Bloom Filter (80-90% of queries, <1ms) - Verified contacts
- **Tier 2**: Identity DB Cache (8-10% of queries, <2ms) - Cached spam scores
- **Tier 3**: Full ML Computation (1-2% of queries, <50ms) - On-demand inference

This strategy reduces average latency from 50ms to <5ms and reduces ML model invocations by 100x.

### 2. Real-time + Batch Hybrid Architecture
- **Flink**: Real-time stream processing for continuous feature aggregation
- **Spark**: Hourly batch jobs for bulk spam score computation
- **Benefit**: Proactive cache warming + on-demand computation when needed

### 3. Specialized Aerospike Databases
- **Identity DB**: Spam score cache (simple key-value, 7-day TTL)
- **Aggregation DB**: Real-time aggregated metrics (Flink writes, high throughput)
- **Cost**: ~$6-8K/month total vs ~$80K/month if using DynamoDB

### 4. Event-Driven Architecture
- **Kafka Topics**: call-logs-events, spam-scores, contact-sync-events, etc.
- **Flink Consumers**: Real-time aggregation
- **Spark Consumers**: Batch processing from Datawarehouse
- **Benefit**: Decoupled services, scalable processing pipeline

### 5. ML Pipeline
- **Training**: Offline on historical data from Dataswarm House
- **Inference**: 
  - Real-time: On cache miss (Identity Service → ML Model)
  - Batch: Hourly Spark jobs (bulk inference for all new call numbers)
- **Features**: Served from Aggregation DB (<2ms latency)

The architecture leverages modern microservices patterns, specialized databases for different use cases, event-driven design with Kafka, real-time stream processing with Flink, batch ML inference with Spark, and intelligent caching strategies to deliver a robust, scalable, and cost-effective caller identification system.

