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
        PENDING: { step: 1, progress: 0 },
        PREPARING: { step: 2, progress: 65 },
        PICKING_UP: { step: 3, progress: 40 },
        DELIVERED: { step: 4, progress: 100 },
    };

    const current = statusMap[status as keyof typeof statusMap] || statusMap.PREPARING;

    return (
        <div className="w-full relative group">
            {/* 1. Main Container: คลีนที่สุด ลบเงาฟุ้ง ลบ Glow ออกทั้งหมด */}
            <div className="w-full h-24 rounded-[16px] bg-slate-900/90 backdrop-blur-xl border border-white/10 flex flex-col justify-center px-6 shadow-2xl overflow-hidden">

                {/* Top Sheen: เหลือแค่ไฮไลท์กระจกบางๆ ไม่มีความเรืองแสง */}
                <div className="absolute top-0 left-0 right-0 h-[40%] bg-gradient-to-b from-white/[0.05] to-transparent pointer-events-none" />

                {/* 2. Top Section: Info Row */}
                <div className="relative z-10 flex items-center justify-between mb-4">
                    <div className="flex items-center gap-3">
                        <span className="text-xl leading-none">🥘</span>
                        <div>
                            <h3 className="text-white font-bold text-[15px] tracking-tight uppercase antialiased">
                                PREPARING ORDER
                            </h3>
                            <p className="text-slate-400 text-[11px] font-medium opacity-80 tracking-widest">
                                Arriving in 12 mins
                            </p>
                        </div>
                    </div>
                    <div className="bg-white/10 border border-white/10 px-3 py-1.5 rounded-2xl">
                        <p className="text-amber-500 font-mono text-[10px] font-black tracking-tighter">#ORD-9921</p>
                    </div>
                </div>

                {/* 3. Bottom Section: Segmented Track */}
                <div className="relative z-10 flex items-center justify-between gap-3 px-1">
                    {steps.map((step, index) => (
                        <React.Fragment key={step.id}>
                            {/* Milestone Icon: Inactive ชัดเจน (slate-500) และ Active (amber-500) ไม่มีเงาเรืองแสง */}
                            <div className="relative">
                                <step.icon
                                    className={`h-5 w-5 transition-all duration-700 ${current.step >= step.threshold
                                        ? "text-amber-500"
                                        : "text-slate-500"
                                        }`}
                                />
                            </div>

                            {/* Segmented Line: แก้สีเพี้ยน ใช้สีพื้นแบบ Flat High Contrast */}
                            {index < steps.length - 1 && (
                                <div className="relative flex-1 h-2 bg-slate-800 rounded-full overflow-visible">

                                    {/* The Liquid Fill: คมชัด ไม่เรืองแสง */}
                                    <div
                                        className={`absolute inset-y-0 left-0 bg-gradient-to-r from-amber-600 via-amber-400 to-yellow-300 rounded-full transition-all duration-1000 ease-[cubic-bezier(0.23,1,0.32,1)] ${current.step > step.threshold ? "w-full" :
                                            current.step === step.threshold ? "w-[var(--seg-progress)]" : "w-0"
                                            }`}
                                        style={{ '--seg-progress': `${current.progress}%` } as React.CSSProperties}
                                    >
                                        {current.step === step.threshold && (
                                            <div
                                                className="absolute right-0 top-0 -translate-y-[115%] translate-x-1/2 text-[20px] animate-bounce pointer-events-none z-20"
                                                style={{ transform: 'translate(50%, -115%) scaleX(-1)' }}
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

