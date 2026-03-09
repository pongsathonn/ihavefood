'use client'

import { Card, CardContent } from '@/components/ui/card'
import { useState, useRef, useEffect } from 'react'
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
  type CarouselApi,
} from '@/components/ui/carousel'
import Autoplay from 'embla-carousel-autoplay'
import Image from 'next/image'

export default function PromotionCard() {
  const plugin = useRef(Autoplay({ delay: 2000, stopOnInteraction: true }))
  const [api, setApi] = useState<CarouselApi>()
  const [current, setCurrent] = useState(0)
  useEffect(() => {
    if (!api) {
      return
    }

    const onSelect = () => {
      setCurrent(api.selectedScrollSnap())
    }

    setCurrent(api.selectedScrollSnap())
    api.on('select', () => {
      setCurrent(api.selectedScrollSnap())
    })

    return () => {
      api.off('select', onSelect)
    }
  }, [api])

  // const SKELETON_IMAGE =
  //   'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/+9fPQAJJAN0f890HAAAAABJRU5ErkJggg=='

  const promoImages = [
    'https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/promotions/promo1.png',
    'https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/promotions/promo2.png',
    'https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/promotions/promo3.png',
    'https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/promotions/promo4.png',
    'https://wadpgqajugnhnkf9.public.blob.vercel-storage.com/promotions/promo5.png',
  ]

  return (
    <Carousel
      plugins={[plugin.current]}
      setApi={setApi}
      className="w-full"
      opts={{
        align: 'center',
        loop: true,
        startIndex: 0,
      }}
      onMouseEnter={plugin.current.stop}
      onMouseLeave={plugin.current.reset}
    >
      <CarouselContent className="items-center py-12">
        {promoImages.map((img, index) => (
          <CarouselItem key={index} className="basis-full sm:basis-1/3">
            <div className="p-2">
              <Card
                className={`
                            relative
                            aspect-video
                            overflow-hidden
                            transition-all duration-500
                            ${
                              index === current
                                ? 'z-10 shadow-2xl ring-2 ring-primary scale-105 md:scale-110 opacity-100'
                                : 'z-0 scale-90 opacity-50'
                            }
                        `}
              >
                <CardContent className="absolute inset-0 flex items-center justify-center p-0">
                  {/*<span className="relative z-10 text-4xl font-semibold text-white">
                    {index + 1}
                  </span>*/}
                  <Image
                    src={img}
                    alt={`promotion ${index + 1}`}
                    fill
                    className="object-cover"
                    sizes="(max-width: 768px) 100vw, (max-width: 1200px) 33vw, 426px"
                    priority={index === 0}
                  />
                </CardContent>
              </Card>
            </div>
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious className="hidden md:flex" />
      <CarouselNext className="hidden md:flex" />
    </Carousel>
  )
}
