---
tags: ["tibrain", "testing", "documentation", "skill"]
scopes: ["code", "tibrain"]
last_updated: 2026-05-22
---
# Computer Vision Integration cho Rust Browser Automation

> **Ngày tạo**: 2026-04-29
> **Mục đích**: Tài liệu kiến thức về tích hợp computer vision vào Rust projects cho browser automation
> **Áp dụng cho**: Donut Browser AI Module

---

## 📚 Tổng Quan

Computer Vision (CV) cho browser automation giúp:
- Phát hiện UI elements từ screenshots (buttons, inputs, links)
- Phân tích layout và accessibility
- Detect visual anomalies
- Generate actionable insights cho AI agents

---

## 🔧 Rust Computer Vision Libraries

### 1. **image** (✅ Đã có trong Donut Browser)
- **Version**: 0.25
- **Mục đích**: Basic image I/O, manipulation
- **Features**: Decode/encode PNG, JPEG, WebP; resize, crop, color conversion
- **Usage**: 
  ```toml
  [dependencies]
  image = "0.25"
  ```

### 2. **imageproc**
- **Mục đích**: Image processing algorithms
- **Features**: Edge detection, filtering, morphological operations
- **Usage**:
  ```toml
  [dependencies]
  imageproc = "0.24"
  ```

### 3. **opencv-rust**
- **Mục đích**: OpenCV bindings cho Rust
- **Features**: Full OpenCV API, ML models, object detection
- **Trade-off**: Heavy dependency, cần OpenCV cài đặt trên system
- **Usage**:
  ```toml
  [dependencies]
  opencv = { version = "0.88", features = ["opencv-4"] }
  ```

### 4. **tflite-rust**
- **Mục đích**: TensorFlow Lite cho on-device ML
- **Features**: Run pre-trained TFLite models
- **Usage**:
  ```toml
  [dependencies]
  tflite = "0.12"
  ```

### 5. **tract**
- **Mục đích**: Neural network inference engine
- **Features**: ONNX, TensorFlow, PyTorch models
- **Performance**: Fast inference, no Python dependency
- **Usage**:
  ```toml
  [dependencies]
  tract = { version = "0.21", features = ["onnx"] }
  ```

### 6. **burn**
- **Mục đích**: Deep learning framework thuần Rust
- **Features**: Training & inference, dynamic graphs
- **Status**: Đang phát triển active
- **Usage**:
  ```toml
  [dependencies]
  burn = { version = "0.13", features = ["train"] }
  ```

---

## 🎯 Integration Patterns cho Browser Automation

### Pattern 1: Heuristic-Based Element Detection (MVP)

**Approach**: Sử dụng image processing algorithms thay vì ML models

**Steps**:
1. **Edge Detection** - Tìm boundaries của UI elements
2. **Color Analysis** - Phân biệt buttons vs background
3. **Shape Detection** - Identify rectangles, circles (buttons, inputs)
4. **Text Detection** - Sử dụng OCR hoặc font analysis

**Pros**:
- ✅ Không cần ML model
- ✅ Nhanh, low latency
- ✅ Easy to debug

**Cons**:
- ❌ Lower accuracy
- ❌ Hard-coded rules
- ❌ Limited cho complex layouts

**Implementation**:
```rust
use image::{ImageBuffer, Rgb};
use imageproc::edges::canny;
use imageproc::contrast::threshold;

pub fn detect_elements_heuristic(screenshot: &[u8]) -> Vec<Element> {
    // 1. Load image
    let img = image::load_from_memory(screenshot)?;
    
    // 2. Edge detection
    let edges = canny(&img, 50.0, 100.0);
    
    // 3. Threshold
    let thresholded = threshold(&edges, 128);
    
    // 4. Find contours (bounding boxes)
    let contours = find_contours(&thresholded);
    
    // 5. Classify elements based on shape/color
    classify_elements(contours)
}
```

---

### Pattern 2: ML-Based Element Detection (Production)

**Approach**: Sử dụng pre-trained object detection models

**Models**:
- **YOLOv8** - Real-time object detection
- **EfficientDet** - Balance speed/accuracy
- **Custom models** - Trained trên UI element datasets

**Steps**:
1. Load model (ONNX/TFLite)
2. Preprocess screenshot (resize, normalize)
3. Run inference
4. Postprocess results (NMS, filtering)
5. Convert to Element structs

**Implementation**:
```rust
use tract_onnx::prelude::*;

pub fn detect_elements_ml(screenshot: &[u8]) -> Result<Vec<Element>, AIError> {
    // 1. Load ONNX model
    let model = tract_onnx::onnx()
        .model_for_path("models/ui_elements.onnx")?
        .into_runnable()?;
    
    // 2. Preprocess image
    let input = preprocess_image(screenshot)?;
    
    // 3. Run inference
    let output = model.run(tvec!(input))?;
    
    // 4. Postprocess
    let detections = postprocess_detections(output)?;
    
    Ok(detections)
}
```

---

### Pattern 3: Hybrid Approach (Recommended)

**Approach**: Kết hợp heuristic + ML

**Strategy**:
1. **Fast path**: Heuristic detection cho common elements
2. **Slow path**: ML detection cho complex/ambiguous cases
3. **Fallback**: Accessibility API nếu available

**Benefits**:
- ✅ Fast cho simple cases
- ✅ Accurate cho complex cases
- ✅ Graceful degradation

---

## 🖼️ Screenshot Analysis Patterns

### Layout Analysis
```rust
pub struct ScreenshotAnalysis {
    pub layout_score: f64,        // Layout quality (0-1)
    pub accessibility_score: f64, // WCAG compliance
    pub anomalies: Vec<String>,  // Visual issues
    pub recommendations: Vec<String>,
}

pub fn analyze_screenshot(screenshot: &[u8]) -> Result<ScreenshotAnalysis, AIError> {
    let img = image::load_from_memory(screenshot)?;
    
    // 1. Detect layout issues
    let layout_issues = detect_layout_issues(&img)?;
    
    // 2. Check accessibility (contrast, spacing)
    let accessibility = check_accessibility(&img)?;
    
    // 3. Identify anomalies
    let anomalies = detect_anomalies(&img)?;
    
    // 4. Generate recommendations
    let recommendations = generate_recommendations(&layout_issues, &accessibility)?;
    
    Ok(ScreenshotAnalysis {
        layout_score: calculate_layout_score(&layout_issues),
        accessibility_score: calculate_accessibility_score(&accessibility),
        anomalies,
        recommendations,
    })
}
```

### Accessibility Metrics
- **Contrast ratio** - WCAG AA/AAA compliance
- **Element spacing** - Touch target size (44x44px min)
- **Text readability** - Font size, line height
- **Color blindness** - Simulate color vision deficiencies

---

## 🔗 Accessibility Tree Parsing

### Approach
Browser accessibility trees (via CDP/Playwright) cung cấp structured data:
- Element roles (button, link, input)
- Labels và descriptions
- Focusable elements
- Screen reader content

### Implementation
```rust
use playwright::Page;

pub async fn parse_accessibility_tree(page: &Page) -> Result<Vec<AccessibleElement>, AIError> {
    // 1. Get accessibility tree via CDP
    let tree = page.accessibility_snapshot().await?;
    
    // 2. Parse tree into structured elements
    let elements = parse_tree_nodes(tree)?;
    
    // 3. Enrich with screenshot coordinates
    let enriched = enrich_with_coordinates(elements, page).await?;
    
    Ok(enriched)
}
```

---

## 🚀 Performance Optimization

### 1. Caching Strategies
- **Screenshot cache** - Cache processed screenshots
- **Detection cache** - Cache element detection results
- **Model warm-up** - Load models at startup

### 2. Parallel Processing
```rust
use rayon::prelude::*;

pub fn detect_elements_parallel(screenshots: Vec<&[u8]>) -> Vec<Vec<Element>> {
    screenshots.par_iter()
        .map(|s| detect_elements(s).unwrap_or_default())
        .collect()
}
```

### 3. Downscaling
- Process thumbnails cho fast detection
- Full resolution cho accurate analysis
- Adaptive resolution dựa trên task complexity

---

## 📊 Best Practices

### 1. Error Handling
```rust
pub fn detect_elements_safe(screenshot: &[u8]) -> Result<Vec<Element>, AIError> {
    // Validate input
    if screenshot.is_empty() {
        return Err(AIError::InvalidInput("Empty screenshot".to_string()));
    }
    
    // Try multiple approaches
    match detect_elements_ml(screenshot) {
        Ok(elements) => Ok(elements),
        Err(e) => {
            log::warn!("ML detection failed: {}, falling back to heuristic", e);
            detect_elements_heuristic(screenshot)
        }
    }
}
```

### 2. Memory Management
- Use streaming cho large images
- Free buffers ngay sau khi dùng
- Limit concurrent image processing

### 3. Logging
- Log detection accuracy
- Track performance metrics
- Monitor model inference time

---

## 🎯 Implementation Roadmap cho Donut Browser

### Phase 1: Foundation (Current)
- ✅ Define data structures (Element, BoundingBox, ElementType)
- ✅ Create module structure (element_detector.rs, screenshot_analyzer.rs)
- 🔄 Implement heuristic-based detection
- ⏳ Implement screenshot analysis
- ⏳ Add accessibility tree parsing

### Phase 2: ML Integration
- ⏳ Choose ML framework (tract hoặc tflite)
- ⏳ Train/fine-tune model cho UI elements
- ⏳ Implement model loading & inference
- ⏳ Add model versioning & updates

### Phase 3: Optimization
- ⏳ Add caching layer
- ⏳ Implement parallel processing
- ⏳ Performance benchmarking
- ⏳ Memory optimization

---

## 📚 References

- **Rust Image Processing**: https://github.com/image-rs/image
- **Tract ONNX**: https://github.com/sonos/tract
- **OpenCV Rust**: https://github.com/twistedfall/opencv-rust
- **Playwright Accessibility**: https://playwright.dev/docs/accessibility

---

## 📝 Lessons Learned

1. **Start simple**: Heuristic-based detection đủ cho MVP
2. **Hybrid approach**: Kết hợp fast + accurate paths
3. **Fallback critical**: Luôn có fallback khi ML fail
4. **Performance matters**: Browser automation cần low latency
5. **Accessibility first**: Accessibility tree thường chính xác hơn CV

---

*Last Updated: 2026-04-29*
