'use client';

import { useState, useRef, useEffect } from 'react';
import { X, ZoomIn, ZoomOut, RotateCw } from 'lucide-react';

interface ImageCropModalProps {
  isOpen: boolean;
  imageFile: File;
  onClose: () => void;
  onCropComplete: (croppedFile: File) => void;
}

export default function ImageCropModal({
  isOpen,
  imageFile,
  onClose,
  onCropComplete,
}: ImageCropModalProps) {
  const [image, setImage] = useState<HTMLImageElement | null>(null);
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [rotation, setRotation] = useState(0);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const CANVAS_SIZE = 400;
  const CROP_SIZE = 300;

  useEffect(() => {
    if (imageFile && isOpen) {
      const img = new Image();
      const objectUrl = URL.createObjectURL(imageFile);

      img.onload = () => {
        setImage(img);
        // Center image initially
        const initialScale = Math.max(
          CROP_SIZE / img.width,
          CROP_SIZE / img.height
        ) * 1.2;
        setScale(initialScale);
        setPosition({ x: 0, y: 0 });
      };

      img.src = objectUrl;

      return () => {
        URL.revokeObjectURL(objectUrl);
      };
    }
  }, [imageFile, isOpen]);

  useEffect(() => {
    if (image && canvasRef.current) {
      drawCanvas();
    }
  }, [image, scale, position, rotation]);

  const drawCanvas = () => {
    const canvas = canvasRef.current;
    if (!canvas || !image) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Clear canvas
    ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE);

    // Save context state
    ctx.save();

    // Draw background (darker overlay)
    ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
    ctx.fillRect(0, 0, CANVAS_SIZE, CANVAS_SIZE);

    // Create circular clip path for crop area
    ctx.beginPath();
    ctx.arc(CANVAS_SIZE / 2, CANVAS_SIZE / 2, CROP_SIZE / 2, 0, Math.PI * 2);
    ctx.clip();

    // Clear the crop area (make it transparent to show the image)
    ctx.clearRect(0, 0, CANVAS_SIZE, CANVAS_SIZE);

    // Move to center and apply transformations
    ctx.translate(CANVAS_SIZE / 2, CANVAS_SIZE / 2);
    ctx.rotate((rotation * Math.PI) / 180);
    ctx.scale(scale, scale);
    ctx.translate(position.x, position.y);

    // Draw image centered
    ctx.drawImage(
      image,
      -image.width / 2,
      -image.height / 2,
      image.width,
      image.height
    );

    ctx.restore();

    // Draw crop circle border
    ctx.strokeStyle = '#3B82F6';
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.arc(CANVAS_SIZE / 2, CANVAS_SIZE / 2, CROP_SIZE / 2, 0, Math.PI * 2);
    ctx.stroke();
  };

  const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
    setIsDragging(true);
    setDragStart({
      x: e.clientX - position.x,
      y: e.clientY - position.y,
    });
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDragging) return;

    setPosition({
      x: e.clientX - dragStart.x,
      y: e.clientY - dragStart.y,
    });
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleTouchStart = (e: React.TouchEvent<HTMLCanvasElement>) => {
    if (e.touches.length !== 1) return;
    setIsDragging(true);
    setDragStart({
      x: e.touches[0].clientX - position.x,
      y: e.touches[0].clientY - position.y,
    });
  };

  const handleTouchMove = (e: React.TouchEvent<HTMLCanvasElement>) => {
    if (!isDragging || e.touches.length !== 1) return;

    setPosition({
      x: e.touches[0].clientX - dragStart.x,
      y: e.touches[0].clientY - dragStart.y,
    });
  };

  const handleZoomIn = () => {
    setScale((prev) => Math.min(prev + 0.1, 3));
  };

  const handleZoomOut = () => {
    setScale((prev) => Math.max(prev - 0.1, 0.5));
  };

  const handleRotate = () => {
    setRotation((prev) => (prev + 90) % 360);
  };

  const handleCrop = async () => {
    const canvas = canvasRef.current;
    if (!canvas || !image) return;

    // Create a new canvas for the cropped image
    const cropCanvas = document.createElement('canvas');
    cropCanvas.width = CROP_SIZE;
    cropCanvas.height = CROP_SIZE;
    const cropCtx = cropCanvas.getContext('2d');
    if (!cropCtx) return;

    // Apply transformations and draw the cropped portion
    cropCtx.save();
    cropCtx.translate(CROP_SIZE / 2, CROP_SIZE / 2);
    cropCtx.rotate((rotation * Math.PI) / 180);
    cropCtx.scale(scale, scale);
    cropCtx.translate(position.x, position.y);
    cropCtx.drawImage(
      image,
      -image.width / 2,
      -image.height / 2,
      image.width,
      image.height
    );
    cropCtx.restore();

    // Convert canvas to blob
    cropCanvas.toBlob(
      (blob) => {
        if (!blob) return;

        // Create new file from blob
        const croppedFile = new File([blob], imageFile.name, {
          type: imageFile.type,
          lastModified: Date.now(),
        });

        onCropComplete(croppedFile);
        onClose();
      },
      imageFile.type,
      0.95
    );
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-75">
      <div className="bg-white rounded-lg shadow-xl p-6 max-w-lg w-full mx-4">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-bold text-gray-900">
            Ajustar Foto de Perfil
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-full transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Canvas */}
        <div ref={containerRef} className="mb-4 flex justify-center">
          <div className="relative" style={{ width: CANVAS_SIZE, height: CANVAS_SIZE }}>
            <canvas
              ref={canvasRef}
              width={CANVAS_SIZE}
              height={CANVAS_SIZE}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseUp}
              onTouchStart={handleTouchStart}
              onTouchMove={handleTouchMove}
              onTouchEnd={handleMouseUp}
              className="border border-gray-300 rounded-lg cursor-move"
              style={{ touchAction: 'none' }}
            />
          </div>
        </div>

        {/* Instructions */}
        <p className="text-sm text-gray-600 text-center mb-4">
          Arraste para reposicionar • Use os controles abaixo para ajustar
        </p>

        {/* Controls */}
        <div className="flex items-center justify-center gap-4 mb-6">
          <button
            onClick={handleZoomOut}
            className="p-2 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors"
            title="Diminuir zoom"
          >
            <ZoomOut className="w-5 h-5 text-gray-700" />
          </button>

          <div className="flex-1 max-w-xs">
            <input
              type="range"
              min="0.5"
              max="3"
              step="0.1"
              value={scale}
              onChange={(e) => setScale(parseFloat(e.target.value))}
              className="w-full"
            />
          </div>

          <button
            onClick={handleZoomIn}
            className="p-2 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors"
            title="Aumentar zoom"
          >
            <ZoomIn className="w-5 h-5 text-gray-700" />
          </button>

          <button
            onClick={handleRotate}
            className="p-2 bg-gray-100 hover:bg-gray-200 rounded-full transition-colors"
            title="Rotacionar 90°"
          >
            <RotateCw className="w-5 h-5 text-gray-700" />
          </button>
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancelar
          </button>
          <button
            onClick={handleCrop}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Aplicar
          </button>
        </div>
      </div>
    </div>
  );
}
