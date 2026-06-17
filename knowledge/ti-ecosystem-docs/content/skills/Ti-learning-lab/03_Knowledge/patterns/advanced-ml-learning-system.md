---
tags: ["tibrain", "documentation", "skill", "router"]
scopes: ["tibrain"]
last_updated: 2026-05-22
---
# Advanced ML Learning System

> **Version**: 1.0.0  
> **Last Updated**: 2026-04-28  
> **Category**: Machine Learning  
> **Language**: Tiếng Việt

---

## 📋 Tổng Quan

Advanced ML Learning System là hệ thống học máy nâng cao cho Ti Router, sử dụng các thuật toán machine learning để tối ưu hóa quyết định routing dựa trên dữ liệu lịch sử.

## 🎯 Mục Tiêu

1. **Dự đoán chi phí** - Dự đoán chi phí cho từng request dựa trên features
2. **Phân loại task** - Phân loại task để chọn model phù hợp
3. **Phân tích sentiment** - Phân tích sentiment của user feedback
4. **Phát hiện anomaly** - Phát hiện các hành vi bất thường

## 🏗️ Architecture

```
LearningSystem
├── AdvancedMLModels
│   ├── CostPredictor (Linear Regression)
│   ├── TaskClassifier (Naive Bayes)
│   ├── SentimentAnalyzer (Simple Sentiment)
│   └── AnomalyDetector (Z-Score)
├── Training Pipeline
│   ├── Data Collection
│   ├── Feature Extraction
│   ├── Model Training
│   └── Model Evaluation
└── Prediction Pipeline
    ├── Feature Extraction
    ├── Model Inference
    └── Result Aggregation
```

## 📊 Các Models

### 1. CostPredictor - Linear Regression

**Mục đích**: Dự đoán chi phí cho request dựa trên features

**Features**:
- Latency (ms)
- Token usage
- Quality score
- Success rate
- Compression ratio
- Retry count

**Training**:
- Thu thập request metrics
- Chuẩn hóa features
- Train linear regression model
- Đánh giá với MSE

**Sử dụng**:
```go
cost := learningSystem.PredictCost(metrics)
```

### 2. TaskClassifier - Naive Bayes

**Mục đích**: Phân loại task để chọn model phù hợp

**Classes**:
- Code generation
- Data analysis
- Text generation
- Question answering
- Translation

**Training**:
- Thu thập labeled examples
- Tính prior probabilities
- Tính likelihood probabilities
- Train Naive Bayes classifier

**Sử dụng**:
```go
taskType := learningSystem.ClassifyTask(prompt)
```

### 3. SentimentAnalyzer - Simple Sentiment

**Mục đích**: Phân tích sentiment của user feedback

**Classes**:
- Positive
- Negative
- Neutral

**Training**:
- Thu thập labeled feedback
- Tính word frequencies
- Train sentiment model

**Sử dụng**:
```go
sentiment := learningSystem.AnalyzeSentiment(feedback)
```

### 4. AnomalyDetector - Z-Score

**Mục đích**: Phát hiện các hành vi bất thường

**Features**:
- Latency
- Cost
- Error rate

**Training**:
- Thu thập metrics
- Tính mean và standard deviation
- Set threshold (thường là 3)

**Sử dụng**:
```go
isAnomaly := learningSystem.DetectAnomaly(metrics)
```

## 🔄 Training Pipeline

### 1. Data Collection

Thu thập metrics từ các requests:
- Request metrics (latency, cost, tokens)
- Task types (prompt, response)
- User feedback (ratings, comments)
- Error logs

### 2. Feature Extraction

Trích xuất features từ raw data:
- Numerical features (latency, cost, tokens)
- Categorical features (task type, provider, model)
- Text features (prompt, feedback)
- Derived features (success rate, compression ratio)

### 3. Model Training

Train các models với features đã trích xuất:
- CostPredictor: Linear regression
- TaskClassifier: Naive Bayes
- SentimentAnalyzer: Simple sentiment
- AnomalyDetector: Z-score

### 4. Model Evaluation

Đánh giá models với metrics:
- CostPredictor: MSE, MAE, R²
- TaskClassifier: Accuracy, Precision, Recall, F1
- SentimentAnalyzer: Accuracy, Confusion Matrix
- AnomalyDetector: Precision, Recall, F1

## 🔮 Prediction Pipeline

### 1. Feature Extraction

Trích xuất features từ input:
- Request metrics
- Task type
- User feedback

### 2. Model Inference

Sử dụng models để dự đoán:
- Cost prediction
- Task classification
- Sentiment analysis
- Anomaly detection

### 3. Result Aggregation

Kết hợp kết quả từ các models:
- Routing decision dựa trên cost prediction
- Model selection dựa trên task classification
- Quality adjustment dựa trên sentiment
- Alert generation dựa trên anomaly detection

## 📈 Performance Metrics

### Training Metrics

| Model | Metric | Target |
|-------|--------|--------|
| CostPredictor | MSE | < 0.01 |
| CostPredictor | R² | > 0.8 |
| TaskClassifier | Accuracy | > 0.9 |
| TaskClassifier | F1 Score | > 0.85 |
| SentimentAnalyzer | Accuracy | > 0.8 |
| AnomalyDetector | Precision | > 0.9 |
| AnomalyDetector | Recall | > 0.8 |

### Inference Metrics

| Metric | Target |
|--------|--------|
| Inference time | < 10ms |
| Memory usage | < 100MB |
| Throughput | > 1000 req/s |

## 🔧 Configuration

```go
advancedMLModels := NewAdvancedMLModels()

// Configure training
advancedMLModels.minSamples = 100
advancedMLModels.learningRate = 0.01
advancedMLModels.epochs = 100

// Configure inference
advancedMLModels.batchSize = 32
advancedMLModels.timeout = 1 * time.Second
```

## 🚀 Usage

### Training Models

```go
// Collect metrics
learningSystem.CollectMetrics(requestMetrics)

// Train models
err := learningSystem.TrainAdvancedMLModels()

// Check performance
performance := learningSystem.GetMLModelPerformance()
```

### Using Predictions

```go
// Predict cost
cost, err := learningSystem.PredictWithML(metrics)

// Classify task
taskType, err := learningSystem.ClassifyTask(prompt)

// Analyze sentiment
sentiment, err := learningSystem.AnalyzeSentiment(feedback)

// Detect anomaly
isAnomaly, err := learningSystem.DetectAnomaly(metrics)
```

## 🎓 Best Practices

### Training

1. **Thu thập đủ data** - Cần ít nhất 100 samples cho training
2. **Chuẩn hóa features** - Scale features về [0, 1]
3. **Cross-validation** - Sử dụng k-fold cross-validation
4. **Regularization** - Thêm regularization để tránh overfitting

### Inference

1. **Batch processing** - Process nhiều requests cùng lúc
2. **Caching** - Cache predictions cho similar requests
3. **Fallback** - Fallback to simple models nếu ML fail
4. **Monitoring** - Monitor prediction accuracy

## 🔍 Troubleshooting

### Training Issues

**Issue**: MSE quá cao
- **Solution**: Thu thập nhiều data hơn, thêm features, regularization

**Issue**: Overfitting
- **Solution**: Regularization, cross-validation, giảm model complexity

**Issue**: Training quá chậm
- **Solution**: Giảm epochs, tăng batch size, sampling data

### Inference Issues

**Issue**: Inference quá chậm
- **Solution**: Batch processing, caching, model quantization

**Issue**: Predictions không chính xác
- **Solution**: Retrain với data mới, thêm features, tune hyperparameters

## 📚 References

- Linear Regression: https://en.wikipedia.org/wiki/Linear_regression
- Naive Bayes: https://en.wikipedia.org/wiki/Naive_Bayes_classifier
- Sentiment Analysis: https://en.wikipedia.org/wiki/Sentiment_analysis
- Anomaly Detection: https://en.wikipedia.org/wiki/Anomaly_detection

---

*Last Updated: 2026-04-28*
