'use client'

import React from "react";
import { Bike, Store, MapPin, ChefHat } from "lucide-react";
import Scooter02Icon from "./scooter-02";

export default function LiveActivity({ status = "PREPARING" }) {
    const steps = [
        { id: "ORDERED", icon: Store, threshold: 1 },
        { id: "PREPARING", icon: ChefHat, threshold: 2 },
        { id: "PICKING_UP", icon: Bike, threshold: 3 },
        { id: "ARRIVED", icon: MapPin, threshold: 4 },
    ];

    const statusMap = {
        PENDING: { step: 1, progress: 0, label: "Order Placed" },
        PREPARING: { step: 2, progress: 65, label: "Preparing Order" },
        PICKING_UP: { step: 3, progress: 40, label: "Rider is Picking Up" },
        DELIVERED: { step: 4, progress: 100, label: "Arrived" },
    };

    const current = statusMap[status as keyof typeof statusMap] || statusMap.PREPARING;

    return (
        <div className="w-full relative">
            {/* Main Container: เปลี่ยนเป็นสีขาว/Slate อ่อนเพื่อให้เข้ากับ Dashboard */}
            <div className="w-full py-6 rounded-[24px] bg-white border border-slate-100 flex flex-col justify-center px-8 shadow-sm overflow-hidden">

                {/* Info Row */}
                <div className="relative z-10 flex items-center justify-between mb-6">
                    <div className="flex items-center gap-4">
                        <div className="w-10 h-10 bg-amber-50 rounded-full flex items-center justify-center text-xl">
                            🥘
                        </div>
                        <div>
                            <h3 className="text-slate-900 font-bold text-[16px] tracking-tight antialiased">
                                {current.label}
                            </h3>
                            <p className="text-slate-400 text-[12px] font-medium tracking-wide">
                                Arriving in 12 mins
                            </p>
                        </div>
                    </div>
                    {/* Badge: ใช้สี Amber บนพื้นอ่อนเพื่อให้ดูแพงขึ้น */}
                    <div className="bg-amber-50 border border-amber-100 px-3 py-1 rounded-full">
                        <p className="text-amber-600 font-mono text-[11px] font-bold">#ORD-9921</p>
                    </div>
                </div>

                {/* Segmented Track */}
                <div className="relative z-10 flex items-center justify-between gap-4 px-1">
                    {steps.map((step, index) => (
                        <React.Fragment key={step.id}>
                            {/* Milestone Icon */}
                            <div className="relative z-20">
                                <div className={`p-2 rounded-full transition-all duration-500 ${current.step >= step.threshold
                                        ? "bg-amber-100 text-amber-600"
                                        : "bg-slate-50 text-slate-300"
                                    }`}>
                                    <step.icon className="h-5 w-5" />
                                </div>
                            </div>

                            {/* Progress Line */}
                            {index < steps.length - 1 && (
                                <div className="relative flex-1 h-[6px] bg-slate-100 rounded-full overflow-visible">
                                    {/* The Liquid Fill: เปลี่ยนเป็น Amber gradient ที่ดูสว่างขึ้น */}
                                    <div
                                        className={`absolute inset-y-0 left-0 bg-gradient-to-r from-amber-400 to-amber-500 rounded-full transition-all duration-1000 ease-[cubic-bezier(0.23,1,0.32,1)] ${current.step > step.threshold ? "w-full" :
                                                current.step === step.threshold ? "w-[var(--seg-progress)]" : "w-0"
                                            }`}
                                        style={{ '--seg-progress': `${current.progress}%` } as React.CSSProperties}
                                    >
                                        {/* Scooter Icon: ปรับสีให้เข้ากับ Theme */}
                                        {current.step === step.threshold && (
                                            <div
                                                className="absolute right-0 top-0 -translate-y-[120%] translate-x-1/2 text-amber-600 animate-bounce pointer-events-none z-30"
                                                style={{ transform: 'translate(50%, -120%) scaleX(-1)' }}
                                            >
                                                <Scooter02Icon />
                                            </div>
                                        )}
                                    </div>
                                </div>
                            )}
                        </React.Fragment>
                    ))}
                </div>
            </div>
        </div>
    );
}