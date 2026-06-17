# Unified Verification Framework Design

## Overview
Comprehensive verification automation framework combining best practices from researched repositories, supporting multiple platforms with unified architecture.

## 🏗️ Architecture Design

### Core Components
```
┌─────────────────────────────────────────────────────────────┐
│                    Unified Verification Framework              │
├─────────────────────────────────────────────────────────────┤
│  Frontend Layer                                              │
│  ├─ Telegram Bot Interface                                  │
│  ├─ Web Dashboard                                           │
│  ├─ API Endpoints                                           │
│  └─ Mobile App (Future)                                     │
├─────────────────────────────────────────────────────────────┤
│  Business Logic Layer                                        │
│  ├─ Verification Engine                                     │
│  ├─ Points System                                           │
│  ├─ User Management                                         │
│  ├─ Admin Controls                                          │
│  └─ Analytics & Monitoring                                  │
├─────────────────────────────────────────────────────────────┤
│  Platform Adapters Layer                                     │
│  ├─ SheerID Adapter                                         │
│  ├─ Email Verification Adapter                              │
│  ├─ Phone Validation Adapter                                │
│  ├─ CAPTCHA Solving Adapter                                 │
│  └─ Custom Platform Adapters                               │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                        │
│  ├─ Database (PostgreSQL/MySQL)                             │
│  ├─ Cache (Redis)                                           │
│  ├─ Message Queue (RabbitMQ/Redis)                          │
│  ├─ Storage (S3/Local)                                      │
│  └─ Monitoring (Prometheus/Grafana)                          │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Technical Implementation

### 1. Core Verification Engine

#### Unified Verification Interface
```python
from abc import ABC, abstractmethod
from typing import Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum

class VerificationStatus(Enum):
    PENDING = "pending"
    PROCESSING = "processing"
    SUCCESS = "success"
    FAILED = "failed"
    TIMEOUT = "timeout"

class VerificationType(Enum):
    STUDENT = "student"
    TEACHER = "teacher"
    MILITARY = "military"
    EMAIL = "email"
    PHONE = "phone"

@dataclass
class VerificationRequest:
    user_id: str
    platform: str
    verification_type: VerificationType
    verification_data: Dict[str, Any]
    priority: int = 1
    metadata: Optional[Dict[str, Any]] = None

@dataclass
class VerificationResult:
    request_id: str
    status: VerificationStatus
    success: bool
    result_data: Optional[Dict[str, Any]]
    error_message: Optional[str]
    processing_time: float
    timestamp: datetime

class VerificationAdapter(ABC):
    """Base adapter for all verification platforms"""
    
    @abstractmethod
    async def verify(self, request: VerificationRequest) -> VerificationResult:
        """Execute verification process"""
        pass
    
    @abstractmethod
    async def validate_request(self, request: VerificationRequest) -> bool:
        """Validate request before processing"""
        pass
    
    @abstractmethod
    def get_platform_config(self) -> Dict[str, Any]:
        """Get platform-specific configuration"""
        pass
    
    @abstractmethod
    async def cleanup_resources(self, request_id: str):
        """Clean up temporary resources"""
        pass
```

#### Verification Engine Implementation
```python
class UnifiedVerificationEngine:
    def __init__(self, config: Dict[str, Any]):
        self.adapters = {}
        self.active_requests = {}
        self.request_queue = asyncio.Queue()
        self.max_concurrent = config.get('max_concurrent', 5)
        self.semaphore = asyncio.Semaphore(self.max_concurrent)
        self.db = DatabaseManager(config['database'])
        self.cache = CacheManager(config['cache'])
        self.metrics = MetricsCollector()
        
    def register_adapter(self, platform: str, adapter: VerificationAdapter):
        """Register a new verification adapter"""
        self.adapters[platform] = adapter
        
    async def submit_verification(self, request: VerificationRequest) -> str:
        """Submit verification request"""
        # Validate request
        adapter = self.adapters.get(request.platform)
        if not adapter:
            raise ValueError(f"Unsupported platform: {request.platform}")
        
        if not await adapter.validate_request(request):
            raise ValueError("Invalid verification request")
        
        # Generate request ID
        request_id = str(uuid.uuid4())
        
        # Store request
        await self.db.store_request(request_id, request)
        
        # Queue for processing
        await self.request_queue.put((request_id, request))
        
        return request_id
    
    async def process_verification_queue(self):
        """Process verification requests from queue"""
        while True:
            try:
                request_id, request = await self.request_queue.get()
                
                async with self.semaphore:
                    await self._process_single_request(request_id, request)
                    
            except Exception as e:
                logger.error(f"Queue processing error: {e}")
                await asyncio.sleep(1)
    
    async def _process_single_request(self, request_id: str, request: VerificationRequest):
        """Process individual verification request"""
        start_time = time.time()
        
        try:
            # Update status to processing
            await self.db.update_request_status(request_id, VerificationStatus.PROCESSING)
            
            # Get adapter
            adapter = self.adapters[request.platform]
            
            # Execute verification
            result = await adapter.verify(request)
            
            # Store result
            await self.db.store_result(request_id, result)
            
            # Record metrics
            processing_time = time.time() - start_time
            self.metrics.record_verification(
                platform=request.platform,
                status=result.status,
                processing_time=processing_time
            )
            
            # Cleanup resources
            await adapter.cleanup_resources(request_id)
            
        except Exception as e:
            # Handle failure
            error_result = VerificationResult(
                request_id=request_id,
                status=VerificationStatus.FAILED,
                success=False,
                result_data=None,
                error_message=str(e),
                processing_time=time.time() - start_time,
                timestamp=datetime.now()
            )
            
            await self.db.store_result(request_id, error_result)
            
        finally:
            # Remove from active requests
            self.active_requests.pop(request_id, None)
    
    async def get_verification_status(self, request_id: str) -> Optional[VerificationResult]:
        """Get verification status"""
        # Check cache first
        cached = await self.cache.get(f"verification:{request_id}")
        if cached:
            return cached
        
        # Check database
        result = await self.db.get_result(request_id)
        if result:
            await self.cache.set(f"verification:{request_id}", result, ttl=3600)
        
        return result
```

### 2. Platform Adapters

#### SheerID Adapter
```python
class SheerIDAdapter(VerificationAdapter):
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.browser = None
        self.session_state = None
        
    async def verify(self, request: VerificationRequest) -> VerificationResult:
        """Execute SheerID verification"""
        try:
            # Extract verification ID from link
            verification_id = self._extract_verification_id(
                request.verification_data['verification_link']
            )
            
            # Generate user data
            user_data = await self._generate_user_data(request.verification_type)
            
            # Create documents
            documents = await self._create_documents(user_data)
            
            # Submit to SheerID
            result = await self._submit_to_sheerid(
                verification_id, user_data, documents
            )
            
            return VerificationResult(
                request_id=request.metadata.get('request_id'),
                status=VerificationStatus.SUCCESS,
                success=True,
                result_data=result,
                error_message=None,
                processing_time=0,  # Will be set by engine
                timestamp=datetime.now()
            )
            
        except Exception as e:
            return VerificationResult(
                request_id=request.metadata.get('request_id'),
                status=VerificationStatus.FAILED,
                success=False,
                result_data=None,
                error_message=str(e),
                processing_time=0,
                timestamp=datetime.now()
            )
    
    async def _generate_user_data(self, verification_type: VerificationType) -> Dict[str, Any]:
        """Generate realistic user data for verification"""
        if verification_type == VerificationType.STUDENT:
            return {
                'first_name': self._generate_name(),
                'last_name': self._generate_name(),
                'email': self._generate_email(),
                'school': self._generate_school(),
                'student_id': self._generate_student_id(),
                'graduation_year': self._generate_graduation_year()
            }
        elif verification_type == VerificationType.TEACHER:
            return {
                'first_name': self._generate_name(),
                'last_name': self._generate_name(),
                'email': self._generate_email(),
                'school': self._generate_school(),
                'employee_id': self._generate_employee_id(),
                'department': self._generate_department(),
                'subject': self._generate_subject()
            }
    
    async def _create_documents(self, user_data: Dict[str, Any]) -> bytes:
        """Create student/teacher ID document"""
        from PIL import Image, ImageDraw, ImageFont
        import io
        
        # Create ID card image
        img = Image.new('RGB', (600, 400), color='white')
        draw = ImageDraw.Draw(img)
        
        # Add text and design
        # ... implementation details ...
        
        # Convert to bytes
        img_bytes = io.BytesIO()
        img.save(img_bytes, format='PNG')
        
        return img_bytes.getvalue()
    
    async def _submit_to_sheerid(self, verification_id: str, user_data: Dict[str, Any], documents: bytes) -> Dict[str, Any]:
        """Submit verification to SheerID"""
        # Initialize browser if needed
        if not self.browser:
            await self._initialize_browser()
        
        # Navigate to verification page
        page = await self.browser.new_page()
        await page.goto(f"https://services.sheerid.com/verify/{verification_id}")
        
        # Fill form
        await self._fill_verification_form(page, user_data)
        
        # Upload documents
        await self._upload_documents(page, documents)
        
        # Submit form
        await page.click('[type="submit"]')
        
        # Wait for result
        await page.wait_for_selector('[data-verification-result]', timeout=30000)
        
        # Extract result
        result = await page.evaluate("""
            () => {
                const element = document.querySelector('[data-verification-result]');
                return element ? element.getAttribute('data-verification-result') : null;
            }
        """)
        
        await page.close()
        
        return {'status': result}
```

#### Email Verification Adapter
```python
class EmailVerificationAdapter(VerificationAdapter):
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.email_client = IMAPClient(config['email'])
        
    async def verify(self, request: VerificationRequest) -> VerificationResult:
        """Extract verification code from email"""
        try:
            # Connect to email
            await self.email_client.connect()
            
            # Get verification codes
            codes = await self._extract_verification_codes(
                request.verification_data['email_subject_pattern']
            )
            
            if codes:
                return VerificationResult(
                    request_id=request.metadata.get('request_id'),
                    status=VerificationStatus.SUCCESS,
                    success=True,
                    result_data={'verification_codes': codes},
                    error_message=None,
                    processing_time=0,
                    timestamp=datetime.now()
                )
            else:
                return VerificationResult(
                    request_id=request.metadata.get('request_id'),
                    status=VerificationStatus.FAILED,
                    success=False,
                    result_data=None,
                    error_message="No verification codes found",
                    processing_time=0,
                    timestamp=datetime.now()
                )
                
        except Exception as e:
            return VerificationResult(
                request_id=request.metadata.get('request_id'),
                status=VerificationStatus.FAILED,
                success=False,
                result_data=None,
                error_message=str(e),
                processing_time=0,
                timestamp=datetime.now()
            )
        finally:
            await self.email_client.disconnect()
    
    async def _extract_verification_codes(self, pattern: str) -> List[str]:
        """Extract verification codes from emails"""
        codes = []
        
        # Search for unread emails
        emails = await self.email_client.search(['UNSEEN'])
        
        for email_id in emails:
            # Fetch email
            email_data = await self.email_client.fetch(email_id, '(RFC822)')
            
            # Parse subject
            subject = email_data.get('subject', '')
            
            # Extract code using pattern
            match = re.search(pattern, subject)
            if match:
                codes.append(match.group(1))
                
                # Mark as read
                await self.email_client.store(email_id, '+FLAGS', '\\Seen')
        
        return codes
```

### 3. Points System

#### Points Management
```python
class PointsSystem:
    def __init__(self, db: DatabaseManager, config: Dict[str, Any]):
        self.db = db
        self.config = config
        self.points_rates = config.get('points_rates', {
            'verify_cost': 1,
            'checkin_reward': 1,
            'invite_reward': 2,
            'register_reward': 1
        })
    
    async def get_user_balance(self, user_id: str) -> int:
        """Get user points balance"""
        balance = await self.db.get_user_balance(user_id)
        return balance or 0
    
    async def add_points(self, user_id: str, amount: int, source: str, metadata: Dict[str, Any] = None) -> bool:
        """Add points to user account"""
        try:
            # Update balance
            await self.db.execute(
                "UPDATE users SET points = points + %s WHERE telegram_id = %s",
                (amount, user_id)
            )
            
            # Log transaction
            await self.db.execute(
                "INSERT INTO points_transactions (user_id, amount, source, metadata, timestamp) VALUES (%s, %s, %s, %s, NOW())",
                (user_id, amount, source, json.dumps(metadata or {}))
            )
            
            return True
            
        except Exception as e:
            logger.error(f"Failed to add points: {e}")
            return False
    
    async def deduct_points(self, user_id: str, amount: int) -> bool:
        """Deduct points from user account"""
        try:
            # Check balance
            current_balance = await self.get_user_balance(user_id)
            if current_balance < amount:
                return False
            
            # Update balance
            await self.db.execute(
                "UPDATE users SET points = points - %s WHERE telegram_id = %s AND points >= %s",
                (amount, user_id, amount)
            )
            
            # Log transaction
            await self.db.execute(
                "INSERT INTO points_transactions (user_id, amount, source, metadata, timestamp) VALUES (%s, %s, %s, %s, NOW())",
                (user_id, -amount, 'verification', json.dumps({}))
            )
            
            return True
            
        except Exception as e:
            logger.error(f"Failed to deduct points: {e}")
            return False
    
    async def process_daily_checkin(self, user_id: str) -> Dict[str, Any]:
        """Process daily check-in reward"""
        # Check if already checked in today
        already_checked = await self.db.execute(
            "SELECT 1 FROM daily_checkins WHERE user_id = %s AND DATE(checkin_date) = CURRENT_DATE",
            (user_id,)
        )
        
        if already_checked:
            return {'success': False, 'message': 'Already checked in today'}
        
        # Add points
        success = await self.add_points(
            user_id, 
            self.points_rates['checkin_reward'], 
            'daily_checkin'
        )
        
        if success:
            # Record check-in
            await self.db.execute(
                "INSERT INTO daily_checkins (user_id, checkin_date) VALUES (%s, CURRENT_DATE)",
                (user_id,)
            )
            
            return {
                'success': True, 
                'points': self.points_rates['checkin_reward'],
                'message': 'Daily check-in successful'
            }
        
        return {'success': False, 'message': 'Failed to process check-in'}
    
    async def process_referral(self, referrer_id: str, referred_id: str) -> Dict[str, Any]:
        """Process referral reward"""
        # Check if referred user is new
        is_new_user = not await self.db.user_exists(referred_id)
        if not is_new_user:
            return {'success': False, 'message': 'User already exists'}
        
        # Check if already referred
        already_referred = await self.db.execute(
            "SELECT 1 FROM referrals WHERE referrer_id = %s AND referred_id = %s",
            (referrer_id, referred_id)
        )
        
        if already_referred:
            return {'success': False, 'message': 'Already referred this user'}
        
        # Add points to referrer
        success = await self.add_points(
            referrer_id,
            self.points_rates['invite_reward'],
            'referral',
            {'referred_user': referred_id}
        )
        
        if success:
            # Record referral
            await self.db.execute(
                "INSERT INTO referrals (referrer_id, referred_id, referral_date) VALUES (%s, %s, CURRENT_DATE)",
                (referrer_id, referred_id)
            )
            
            return {
                'success': True,
                'points': self.points_rates['invite_reward'],
                'message': 'Referral reward granted'
            }
        
        return {'success': False, 'message': 'Failed to process referral'}
```

### 4. Telegram Bot Interface

#### Bot Implementation
```python
class UnifiedVerificationBot:
    def __init__(self, token: str, engine: UnifiedVerificationEngine, points_system: PointsSystem):
        self.token = token
        self.engine = engine
        self.points_system = points_system
        self.application = Application.builder().token(token).build()
        self.setup_handlers()
    
    def setup_handlers(self):
        """Setup bot command handlers"""
        # User commands
        self.application.add_handler(CommandHandler("start", self.start_command))
        self.application.add_handler(CommandHandler("help", self.help_command))
        self.application.add_handler(CommandHandler("balance", self.balance_command))
        self.application.add_handler(CommandHandler("checkin", self.checkin_command))
        self.application.add_handler(CommandHandler("invite", self.invite_command))
        self.application.add_handler(CommandHandler("use", self.use_key_command))
        
        # Verification commands
        self.application.add_handler(CommandHandler("verify", self.verify_command))
        self.application.add_handler(CommandHandler("verify2", self.verify2_command))
        self.application.add_handler(CommandHandler("verify3", self.verify3_command))
        self.application.add_handler(CommandHandler("verify4", self.verify4_command))
        self.application.add_handler(CommandHandler("verify5", self.verify5_command))
        
        # Admin commands
        self.application.add_handler(CommandHandler("admin", self.admin_command))
        
        # URL handler
        self.application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, self.url_handler))
    
    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /start command"""
        user_id = str(update.effective_user.id)
        
        # Register user if not exists
        if not await self.engine.db.user_exists(user_id):
            await self.engine.db.create_user(user_id, update.effective_user.username)
            
            # Add registration bonus
            await self.points_system.add_points(
                user_id,
                self.points_system.points_rates['register_reward'],
                'registration'
            )
        
        welcome_text = """
🎉 **Welcome to Unified Verification Bot!**

📋 **Available Commands:**
/start - Show this message
/balance - Check your points balance
/checkin - Daily check-in (+1 point)
/invite - Generate referral link
/verify <link> - Gemini One Pro verification
/verify2 <link> - ChatGPT Teacher verification
/verify3 <link> - Spotify Student verification
/verify4 <link> - Bolt.new Teacher verification
/verify5 <link> - YouTube Student verification

💰 **Points System:**
• Daily check-in: +1 point
• Referrals: +2 points per person
• Verification cost: 1 point

🔗 **How to Use:**
1. Visit verification page
2. Copy the verification link
3. Send to bot with appropriate command

❓ **Need Help?** Use /help for detailed instructions
        """
        
        await update.message.reply_text(welcome_text, parse_mode='Markdown')
    
    async def verify_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle Gemini One Pro verification"""
        await self._handle_verification(update, 'gemini', VerificationType.TEACHER)
    
    async def verify2_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle ChatGPT Teacher verification"""
        await self._handle_verification(update, 'chatgpt', VerificationType.TEACHER)
    
    async def verify3_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle Spotify Student verification"""
        await self._handle_verification(update, 'spotify', VerificationType.STUDENT)
    
    async def verify4_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle Bolt.new Teacher verification"""
        await self._handle_verification(update, 'boltnew', VerificationType.TEACHER)
    
    async def verify5_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle YouTube Student verification"""
        await self._handle_verification(update, 'youtube', VerificationType.STUDENT)
    
    async def _handle_verification(self, update: Update, platform: str, verification_type: VerificationType):
        """Handle verification request"""
        user_id = str(update.effective_user.id)
        
        # Check if user provided verification link
        if not context.args:
            await update.message.reply_text(
                "❌ Please provide verification link\n\n"
                "Example: /verify https://services.sheerid.com/verify/xxx/?verificationId=yyy"
            )
            return
        
        verification_link = context.args[0]
        
        # Validate link format
        if not self._is_valid_verification_link(verification_link):
            await update.message.reply_text(
                "❌ Invalid verification link format\n\n"
                "Please provide a valid SheerID verification link"
            )
            return
        
        # Check user balance
        balance = await self.points_system.get_user_balance(user_id)
        if balance < self.points_system.points_rates['verify_cost']:
            await update.message.reply_text(
                f"❌ Insufficient points\n\n"
                f"Required: {self.points_system.points_rates['verify_cost']} points\n"
                f"Current balance: {balance} points\n\n"
                f"Use /checkin to earn more points"
            )
            return
        
        # Deduct points
        points_deducted = await self.points_system.deduct_points(
            user_id, 
            self.points_system.points_rates['verify_cost']
        )
        
        if not points_deducted:
            await update.message.reply_text("❌ Failed to deduct points")
            return
        
        # Submit verification request
        try:
            request = VerificationRequest(
                user_id=user_id,
                platform=platform,
                verification_type=verification_type,
                verification_data={'verification_link': verification_link},
                metadata={'request_id': str(uuid.uuid4())}
            )
            
            request_id = await self.engine.submit_verification(request)
            
            await update.message.reply_text(
                "🔄 **Verification Submitted**\n\n"
                f"Request ID: `{request_id}`\n"
                f"Platform: {platform}\n"
                f"Cost: {self.points_system.points_rates['verify_cost']} points\n\n"
                "⏱️ Processing usually takes 3-5 minutes\n"
                "🔔 You'll be notified when complete"
            )
            
        except Exception as e:
            # Refund points on error
            await self.points_system.add_points(
                user_id,
                self.points_system.points_rates['verify_cost'],
                'refund',
                {'error': str(e)}
            )
            
            await update.message.reply_text(f"❌ Failed to submit verification: {str(e)}")
    
    def _is_valid_verification_link(self, link: str) -> bool:
        """Validate verification link format"""
        pattern = r'https://services\.sheerid\.com/verify/[^/]+\?verificationId=[^&]+'
        return re.match(pattern, link) is not None
```

## 🚀 Implementation Plan

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Set up project structure
- [ ] Implement database schema
- [ ] Create base verification engine
- [ ] Set up caching and messaging
- [ ] Implement basic error handling

### Phase 2: Platform Adapters (Week 3-4)
- [ ] Implement SheerID adapter
- [ ] Create email verification adapter
- [ ] Add phone validation adapter
- [ ] Implement CAPTCHA solving adapter
- [ ] Add adapter registry

### Phase 3: User Interface (Week 5-6)
- [ ] Build Telegram bot interface
- [ ] Implement points system
- [ ] Add user management
- [ ] Create admin controls
- [ ] Add monitoring and analytics

### Phase 4: Advanced Features (Week 7-8)
- [ ] Implement web dashboard
- [ ] Add API endpoints
- [ ] Create reseller program
- [ ] Add advanced security features
- [ ] Performance optimization

### Phase 5: Production Deployment (Week 9-10)
- [ ] Set up production infrastructure
- [ ] Implement monitoring and alerting
- [ ] Add backup and recovery
- [ ] Create documentation
- [ ] Performance testing

## 📊 Technical Specifications

### Database Schema
```sql
-- Users table
CREATE TABLE users (
    telegram_id VARCHAR(50) PRIMARY KEY,
    username VARCHAR(100),
    points INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_blocked BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE
);

-- Verification requests
CREATE TABLE verification_requests (
    request_id VARCHAR(50) PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(telegram_id),
    platform VARCHAR(50),
    verification_type VARCHAR(20),
    status VARCHAR(20),
    verification_data JSON,
    result_data JSON,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    processing_time FLOAT
);

-- Points transactions
CREATE TABLE points_transactions (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(telegram_id),
    amount INTEGER,
    source VARCHAR(50),
    metadata JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Daily checkins
CREATE TABLE daily_checkins (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) REFERENCES users(telegram_id),
    checkin_date DATE,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, checkin_date)
);

-- Referrals
CREATE TABLE referrals (
    id SERIAL PRIMARY KEY,
    referrer_id VARCHAR(50) REFERENCES users(telegram_id),
    referred_id VARCHAR(50) REFERENCES users(telegram_id),
    referral_date DATE,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(referrer_id, referred_id)
);

-- API keys
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    key_value VARCHAR(100) UNIQUE,
    user_id VARCHAR(50) REFERENCES users(telegram_id),
    points INTEGER,
    max_uses INTEGER,
    used_count INTEGER DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Configuration
```yaml
# config.yaml
app:
  name: "Unified Verification Framework"
  version: "1.0.0"
  debug: false

database:
  type: "postgresql"
  host: "localhost"
  port: 5432
  name: "verification_db"
  user: "verification_user"
  password: "verification_password"

cache:
  type: "redis"
  host: "localhost"
  port: 6379
  db: 0

verification:
  max_concurrent: 5
  timeout: 300
  retry_attempts: 3
  
points:
  verify_cost: 1
  checkin_reward: 1
  invite_reward: 2
  register_reward: 1

telegram:
  bot_token: "YOUR_BOT_TOKEN"
  admin_user_id: "YOUR_ADMIN_ID"
  channel_username: "your_channel"

platforms:
  sheerid:
    program_ids:
      gemini: "PROGRAM_ID_GEMINI"
      chatgpt: "PROGRAM_ID_CHATGPT"
      spotify: "PROGRAM_ID_SPOTIFY"
      boltnew: "PROGRAM_ID_BOLTNEW"
      youtube: "PROGRAM_ID_YOUTUBE"
  
  email:
    imap_server: "imap.gmail.com"
    smtp_server: "smtp.gmail.com"
    
  phone:
    hlr_api_key: "YOUR_HLR_API_KEY"
    
  captcha:
    api_key: "YOUR_CAPTCHA_API_KEY"
```

## 🔒 Security Considerations

### Authentication & Authorization
- JWT token-based authentication
- Role-based access control
- API rate limiting
- Input validation and sanitization

### Data Protection
- Encryption of sensitive data
- GDPR compliance
- Data retention policies
- Audit logging

### Infrastructure Security
- HTTPS enforcement
- Database encryption
- Backup encryption
- Network security groups

## 📈 Monitoring & Analytics

### Metrics Collection
- Verification success rates
- Processing times
- User engagement
- System performance

### Alerting
- Failed verifications
- System errors
- Performance degradation
- Security events

### Reporting
- Daily/weekly reports
- User analytics
- Revenue tracking
- System health

---

**Status**: ✅ Design Complete  
**Next Steps**: Begin implementation with core infrastructure  
**Priority**: High - Start with database setup and basic verification engine
