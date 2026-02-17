import { Card, CardContent } from "@/components/ui/card"
import { useState, useRef, useEffect } from "react";
import {
    Carousel,
    CarouselContent,
    CarouselItem,
    CarouselNext,
    CarouselPrevious,
    type CarouselApi,
} from "@/components/ui/carousel"
import Autoplay from "embla-carousel-autoplay"
import Image from "next/image";

export const PromotionCard = () => {

    const plugin = useRef(
        Autoplay({ delay: 2000, stopOnInteraction: true })
    )

    const [api, setApi] = useState<CarouselApi>()
    const [current, setCurrent] = useState(0)
    useEffect(() => {
        if (!api) {
            return
        }
        setCurrent(api.selectedScrollSnap())
        api.on("select", () => {
            setCurrent(api.selectedScrollSnap())
        })
    }, [api])

    const FAKE_GRAY_IMAGE = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/+9fPQAJJAN0f890HAAAAABJRU5ErkJggg==";

    return (
        <Carousel
            plugins={[plugin.current]}
            setApi={setApi}
            className="w-full"
            opts={{
                align: "center",
                loop: true,
                startIndex: 0,
            }}
            onMouseEnter={plugin.current.stop}
            onMouseLeave={plugin.current.reset}
        >
            <CarouselContent className="items-center py-12">
                {Array.from({ length: 10 }).map((_, index) => (
                    <CarouselItem key={index} className="basis-full md:basis-1/3">
                        <div className="p-2">
                            <Card
                                className={`
                            relative
                            aspect-video 
                            overflow-hidden
                            transition-all duration-500
                            ${index === current
                                        ? "z-10 shadow-2xl ring-2 ring-primary scale-105 md:scale-110 opacity-100"
                                        : "z-0 scale-90 opacity-50"}
                        `}
                            >
                                <CardContent className="absolute inset-0 flex items-center justify-center p-0">
                                    <span className="relative z-10 text-4xl font-semibold text-white">
                                        {index + 1}
                                    </span>
                                    <Image
                                        src={FAKE_GRAY_IMAGE}
                                        alt="test"
                                        fill
                                        className="object-cover"
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