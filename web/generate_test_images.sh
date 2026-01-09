#!/bin/bash

# 创建测试图片目录结构
echo "📁 创建图片目录..."

mkdir -p images/traffic/bus
mkdir -p images/traffic/car
mkdir -p images/traffic/bike
mkdir -p images/traffic/light

echo "✅ 目录创建完成！"
echo ""
echo "📝 目录结构："
echo "images/"
echo "└── traffic/"
echo "    ├── bus/      # 存放公交车图片 (bus_1.jpg, bus_2.jpg...)"
echo "    ├── car/      # 存放汽车图片 (car_1.jpg, car_2.jpg...)"
echo "    ├── bike/     # 存放自行车图片 (bike_1.jpg, bike_2.jpg...)"
echo "    └── light/    # 存放红绿灯图片 (light_1.jpg, light_2.jpg...)"
echo ""
echo "📌 使用说明："
echo "1. 将对应的图片文件放入各自的目录"
echo "2. 图片命名格式：类型_编号.jpg (例如: bus_1.jpg, car_2.jpg)"
echo "3. 支持的格式：.jpg, .png, .jpeg"
echo "4. 推荐尺寸：200x150 像素"
echo ""
echo "💡 提示：你可以从网上下载或使用AI生成这些图片"
