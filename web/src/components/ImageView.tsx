import { useState, useEffect } from 'react';
import { apiClient } from '../services/api';

interface ImageViewProps {
  bucketId: string;
  fileId: string;
  alt: string;
  className?: string;
  onError?: () => void;
}

export default function ImageView({ 
  bucketId, 
  fileId, 
  alt, 
  className,
  onError 
}: ImageViewProps) {
  const [imageSrc, setImageSrc] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    const loadImage = async () => {
      try {
        setLoading(true);
        setError(false);
        
        // Always use authenticated download for consistency
        const blob = await apiClient.downloadFile(bucketId, fileId);
        
        // Verify the blob is an image
        if (blob.type.startsWith('image/')) {
          const objectUrl = URL.createObjectURL(blob);
          setImageSrc(objectUrl);
        } else {
          throw new Error(`File is not an image. Content type: ${blob.type}`);
        }
      } catch (err) {
        console.error('Failed to load image:', err);
        setError(true);
        onError?.();
      } finally {
        setLoading(false);
      }
    };

    loadImage();

    // Cleanup function to revoke blob URLs to prevent memory leaks
    return () => {
      if (imageSrc && imageSrc.startsWith('blob:')) {
        URL.revokeObjectURL(imageSrc);
      }
    };
  }, [bucketId, fileId, onError]);

  // Cleanup blob URL when component unmounts or image changes
  useEffect(() => {
    return () => {
      if (imageSrc && imageSrc.startsWith('blob:')) {
        URL.revokeObjectURL(imageSrc);
      }
    };
  }, [imageSrc]);

  if (loading) {
    return (
      <div className={`${className} flex items-center justify-center bg-dark-700`}>
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (error || !imageSrc) {
    return (
      <div className={`${className} flex items-center justify-center bg-dark-700`}>
        <svg className="h-8 w-8 text-dark-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </div>
    );
  }

  return (
    <img
      src={imageSrc}
      alt={alt}
      className={className}
      onError={() => {
        setError(true);
        onError?.();
      }}
    />
  );
}