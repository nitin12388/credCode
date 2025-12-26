# Truecaller System - Technical Design Document

## 1. System Overview

Truecaller is a caller identification and spam detection system that allows users to:
- Identify unknown callers
- Block spam calls
- Search for phone numbers
- Sync and manage contacts
- Report spam numbers

## 2. Architecture Components

### 2.1 Core Services

#### User Service
**Responsibilities:**
- User registration and authentication
- Contact synchronization
- User profile management
- Device management

**APIs:**
- `POST /api/v1/login` - User authentication
- `POST /api/v1/contact-sync` - Sync contacts from user's phone
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile

#### Search Service
**Responsibilities:**
- Search phone numbers in the database
- Return caller information
- Aggregate data from multiple sources

**APIs:**
- `GET /api/v1/search?phone={phoneNumber}` - Search by phone number
- `GET /api/v1/search?name={name}` - Search by name

#### Identity Service
**Responsibilities:**
- Manage user identity
- Device verification
- Token generation and validation
- Session management

**APIs:**
- `POST /api/v1/identity/verify` - Verify user identity
- `POST /api/v1/identity/token` - Generate access token
- `DELETE /api/v1/identity/logout` - Logout user

#### Call Intercept Service
**Responsibilities:**
- Real-time call interception
- Spam detection using ML models
- Caller ID lookup
- Call event tracking

**APIs:**
- `POST /api/v1/call-intercept` - Process incoming call
- `POST /api/v1/report-spam` - Report spam number

### 2.2 Data Layer

#### User Database (PostgreSQL)

**Database Choice: PostgreSQL**

**Rationale:**
- ACID compliance for transactional consistency
- Excellent support for complex queries and joins
- Mature ecosystem with proven reliability
- Good performance for structured relational data
- Strong community support and tooling

**Schema:**
```
User Table:
- phoneNumber (PK, indexed)
- userDeviceId
- deviceId
- name
- email
- registrationDate
- lastSync
- contactList (array of contacts)
```

**Contact Structure:**
```
- contactPhoneNumber
- contactName
- addedDate
```

**Scaling Considerations:**
- Vertical scaling for initial growth
- Read replicas for read-heavy workloads
- Partitioning by phone number hash for horizontal scaling
- Connection pooling (PgBouncer) for efficient connection management

#### Search Database
**Optimized for:**
- Fast phone number lookups
- Name-based searches
- Fuzzy matching
- Inverted index for name searches

**Schema:**
```
Search Index:
- phoneNumber (PK)
- names (array of names from different sources)
- spamScore
- category (business/personal)
- verificationStatus
```

#### Identity Database (Aerospike)

**Database Choice: Aerospike**

**Rationale:**
- Ultra-low latency reads/writes (sub-millisecond)
- Excellent for session management and identity verification
- Cost-effective compared to alternatives
- Strong consistency with high availability
- Built-in TTL support for session expiration

**Cost Analysis:**
- **Aerospike**: 7-node cluster with i4i.xlarge instances
    - Estimated cost: ~$3,000/month
    - Provides high throughput and low latency

- **DynamoDB (Rejected)**:
    - Estimated cost: ~$40,000/month
    - Too expensive for our use case
    - Less predictable pricing with on-demand model

**Capacity Estimations:**
- **Total Rows**: 3 billion records
- **Row Size**: 0.5 KB per record
- **Total Storage**: 3B × 0.5 KB = 1.5 TB
- **With Replication (RF=2)**: ~3 TB
- **With Overhead**: ~3.5 TB total

**Schema:**
```
Namespace: identity
Set: sessions

Key: sessionToken/deviceId
Value: {
  userId
  phoneNumber
  deviceId
  expiresAt
  refreshToken
  lastActivity
}
```

**Configuration:**
- Replication Factor: 2 (for high availability)
- Storage: SSD (for low latency)
- Memory: Hybrid (index in memory, data on SSD)
- TTL: 24 hours for sessions

### 2.3 Caching Layer

**Cache Strategy:**
- **Hot Data**: Frequently searched numbers (TTL: 1 hour)
- **User Sessions**: Active user sessions (TTL: 24 hours)
- **Search Results**: Recent search results (TTL: 30 minutes)

**Cache Implementation:**
- Redis for distributed caching
- LRU eviction policy
- Write-through for critical data

### 2.4 Message Queue (Kafka)

**Topics:**
1. **contact-sync-events**: Contact sync events for async processing
2. **call-events**: Call intercept events for analytics
3. **spam-reports**: User spam reports for ML training
4. **search-events**: Search events for analytics

**Use Cases:**
- Decouple services
- Async processing of heavy operations
- Event sourcing for analytics
- ML model training data

### 2.5 ML Models Framework

**Models:**
1. **Spam Detection Model**
    - Input: Call patterns, user reports, number characteristics
    - Output: Spam probability score (0-1)

2. **Name Prediction Model**
    - Input: Phone number patterns, historical data
    - Output: Predicted name/category

3. **Aggregation Model**
    - Combines multiple data sources
    - Confidence scoring

### 2.6 Feature Store (Aerospike)

**Database Choice: Aerospike**

**Rationale:**
- Extremely low latency for real-time ML feature serving (<1ms p99)
- High throughput for concurrent feature lookups
- Efficient storage for large-scale feature vectors
- Strong consistency for feature updates
- Cost-effective for high-volume operations

**Purpose:**
- Store pre-computed features for ML models
- Real-time feature serving
- Historical feature tracking
- Fast batch feature retrieval

**Features Stored:**
- Call frequency patterns (last 7 days, 30 days)
- Spam report count and scores
- User engagement metrics
- Number age and history
- Call duration statistics
- Geographic information
- Time-based patterns

**Schema:**
```
Namespace: features
Set: phone_features

Key: phoneNumber
Value: {
  callFrequency7d: int
  callFrequency30d: int
  spamReportCount: int
  spamScore: float
  lastCallTimestamp: timestamp
  avgCallDuration: float
  geographicRegion: string
  accountAge: int
  engagementScore: float
  features: map<string, float>  // flexible feature map
}
```

**Performance:**
- Read Latency: <1ms (p99)
- Write Latency: <2ms (p99)
- Throughput: 100K+ reads/sec per node

### 2.7 Data Warehouse (Dataswarm House)

**Purpose:**
- Long-term storage of historical data
- Analytics and reporting
- ML model training
- Business intelligence

**Data:**
- User activity logs
- Call history
- Search patterns
- Spam reports

### 2.8 Batch Processing (Spark/Flink)

**Jobs:**
1. **Daily Spam Score Update**
    - Aggregate spam reports
    - Update spam scores
    - Feed ML models

2. **Contact Graph Building**
    - Build social graph from contacts
    - Identify communities
    - Improve search relevance

3. **Feature Computation**
    - Compute batch features for ML
    - Update feature store

## 3. Data Flow

### 3.1 User Registration Flow
1. User installs app and provides phone number
2. User Service validates and creates account
3. Identity Service generates session token
4. User DB stores user information
5. Cache stores session data

### 3.2 Contact Sync Flow
1. User grants contact permission
2. App sends contacts to User Service
3. User Service validates and stores in User DB
4. Event published to Kafka (contact-sync-events)
5. Aggregator processes contacts asynchronously
6. Search DB updated with new name-number mappings

### 3.3 Search Flow
1. User searches for a phone number
2. API Gateway routes to Search Service
3. Search Service checks Cache first
4. If cache miss, queries Search DB
5. Results aggregated from multiple sources
6. Cache updated with results
7. Search event logged to Kafka

### 3.4 Call Intercept Flow
1. Incoming call triggers app
2. Call Intercept Service receives phone number
3. Check Cache for caller information
4. If cache miss, query User DB and Search DB
5. ML Model evaluates spam probability
6. Display caller information to user
7. Log call event to Kafka

### 3.5 Spam Report Flow
1. User reports a number as spam
2. Event sent to Call Intercept Service
3. Update Search DB with spam report
4. Publish to Kafka (spam-reports)
5. ML model retraining triggered (batch)
6. Feature Store updated

## 4. Scalability Considerations

### 4.1 Horizontal Scaling
- All services are stateless and can scale horizontally
- Load balancer distributes traffic across instances
- Database sharding based on phone number hash

### 4.2 Database Sharding
**Sharding Strategy:**
- Shard key: Hash of phone number
- Number of shards: Based on data volume
- Consistent hashing for even distribution

### 4.3 Caching Strategy
- Multi-level caching (L1: Local, L2: Redis)
- Cache warming for popular numbers
- Cache invalidation on data updates

### 4.4 Rate Limiting
- Per-user rate limits to prevent abuse
- API-level throttling
- DDoS protection at API Gateway

## 5. Security & Privacy

### 5.1 Data Privacy
- Contact data encrypted at rest
- Minimal PII storage
- User consent for data usage
- GDPR compliance

### 5.2 Authentication & Authorization
- JWT tokens for API authentication
- Device-based authentication
- Session management
- Role-based access control

### 5.3 Data Anonymization
- Contact data anonymized before ML training
- Aggregated analytics without PII
- User opt-out options

## 6. Performance Metrics

### 6.1 SLAs
- Search API: < 100ms (p99)
- Call Intercept: < 50ms (p99)
- Contact Sync: < 5s for 1000 contacts
- Uptime: 99.9%

### 6.2 Monitoring
- Request latency tracking
- Error rate monitoring
- Cache hit ratio
- Database query performance
- Kafka lag monitoring

## 7. API Gateway & Load Balancer

### 7.1 Responsibilities
- Request routing
- SSL termination
- Rate limiting
- Authentication
- Load balancing across service instances

### 7.2 Load Balancing Strategy
- Round-robin for stateless services
- Least connections for heavy operations
- Geographic routing for reduced latency

## 8. Technology Stack

- **Language**: Go (for high performance and concurrency)
- **Databases**:
    - **PostgreSQL** (User DB) - User profiles and contacts
    - **Aerospike** (Identity DB) - Sessions, tokens, device management
    - **Aerospike** (Feature Store) - ML features for real-time serving
    - **Elasticsearch** (Search DB) - Phone number and name searches
    - **Redis** (Cache) - Hot data caching layer
- **Message Queue**: Apache Kafka
- **ML Framework**: TensorFlow/PyTorch
- **Batch Processing**: Apache Spark/Flink
- **Containerization**: Docker/Kubernetes
- **API Gateway**: Kong/Envoy
- **Monitoring**: Prometheus, Grafana
- **Logging**: ELK Stack
- **Tracing**: Jaeger/Zipkin

## 9. Implementation Considerations

### 9.1 Data Consistency
- Eventual consistency for non-critical data
- Strong consistency for user authentication
- Saga pattern for distributed transactions

### 9.2 Fault Tolerance
- Circuit breakers for external dependencies
- Retry mechanisms with exponential backoff
- Fallback mechanisms for service failures
- Dead letter queues for failed messages

### 9.3 Testing Strategy
- Unit tests for business logic
- Integration tests for service interactions
- Load testing for scalability validation
- Chaos engineering for resilience testing

## 10. Database Cost Analysis & Justification

### 10.1 Identity Database: Aerospike vs DynamoDB

#### Aerospike (Selected)
**Configuration:**
- Instance Type: i4i.xlarge (8 vCPUs, 32 GB RAM, 1x 937 GB NVMe SSD)
- Cluster Size: 7 nodes
- Replication Factor: 2
- Region: US-East-1

**Cost Breakdown:**
- Instance Cost: 7 nodes × $0.425/hour = $2.975/hour
- Monthly Cost: $2.975 × 730 hours = ~$2,172/month
- Data Transfer: ~$500/month
- **Total: ~$3,000/month**

**Capacity:**
- Storage per node: 937 GB × 7 = 6.5 TB raw
- Usable storage (RF=2): ~3.25 TB
- Current requirement: 1.5 TB (with growth room)
- IOPS: 250K+ IOPS per node
- Latency: <1ms (p99)

**Benefits:**
- Predictable, linear pricing
- Excellent performance characteristics
- No hidden costs or surprise bills
- Simple capacity planning

#### DynamoDB (Rejected)
**Estimated Configuration:**
- On-Demand pricing for unpredictable workload
- 3 billion items × 0.5 KB = 1.5 TB storage
- Estimated reads: 100,000 reads/sec
- Estimated writes: 10,000 writes/sec

**Cost Breakdown:**
- Storage: 1.5 TB × $0.25/GB = $375/month
- Read Request Units: 100K RPS × 0.5 RCU × $0.25/million = ~$10,800/month
- Write Request Units: 10K WPS × 1 WCU × $1.25/million = ~$10,800/month
- Data Transfer: ~$1,000/month
- Backup & Point-in-time recovery: ~$2,000/month
- Global Tables (if needed): Additional 2x cost
- **Total: ~$40,000/month (could spike higher)**

**Issues:**
- Unpredictable costs with traffic spikes
- Complex pricing model (RCU/WCU)
- Higher latency (single-digit milliseconds)
- Potential throttling issues
- Cost scales non-linearly with traffic

### 10.2 Feature Store Database: Aerospike

**Rationale for Aerospike:**
- ML features require <1ms latency for real-time inference
- High read throughput (100K+ reads/sec)
- Efficient storage for feature vectors
- Consistent performance under load
- Cost-effective for high-volume operations

**Configuration:**
- Similar to Identity DB setup
- Can be co-located or separate cluster based on load
- Estimated Cost: ~$3,000-5,000/month depending on scale

**Alternatives Considered:**
- **Redis**: Too expensive for large datasets, memory-only
- **Cassandra**: Higher operational complexity, slower
- **DynamoDB**: Cost prohibitive as shown above

### 10.3 User Database: PostgreSQL

**Rationale:**
- Structured relational data with complex relationships
- ACID compliance for user data integrity
- Excellent query performance with proper indexing
- Mature ecosystem and tooling
- Cost-effective with predictable pricing

**Estimated Configuration:**
- RDS PostgreSQL (db.r6g.4xlarge)
- Multi-AZ deployment for high availability
- Estimated Cost: ~$2,000-3,000/month
- Read replicas for scaling reads: +$1,000-2,000/month

### 10.4 Total Database Infrastructure Cost Summary

| Database | Technology | Use Case | Monthly Cost |
|----------|-----------|----------|--------------|
| Identity DB | Aerospike (7 nodes) | Sessions, Tokens | ~$3,000 |
| Feature Store | Aerospike (5-7 nodes) | ML Features | ~$3,000-5,000 |
| User DB | PostgreSQL RDS | User Profiles | ~$3,000-5,000 |
| Search DB | Elasticsearch | Search Indices | ~$4,000-6,000 |
| Cache | Redis | Hot Data Cache | ~$1,000-2,000 |
| **TOTAL** | | | **~$14,000-21,000/month** |

**Cost Savings:**
- By choosing Aerospike over DynamoDB for Identity DB: **$37,000/month saved**
- By choosing Aerospike for Feature Store over DynamoDB: **~$35,000/month saved**
- **Total Monthly Savings: ~$72,000/month or $864,000/year**

## 11. Future Enhancements

1. **Video Call Support**: Identify video callers
2. **Message Spam Detection**: Extend to SMS/messaging apps
3. **Community Features**: User communities and forums
4. **Advanced Analytics**: Call pattern insights for users
5. **International Support**: Multi-language and region-specific features
6. **Edge Computing**: Deploy lightweight services at edge locations for reduced latency
