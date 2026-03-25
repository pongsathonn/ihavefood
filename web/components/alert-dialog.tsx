import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"

export function AlertConfirm({
    isItemAdd,
    handleOrderPlace,
}: {
    isItemAdd: boolean | undefined
    handleOrderPlace: () => Promise<void>
}) {
    return (
        <AlertDialog>
            <AlertDialogTrigger asChild>
                <Button
                    size="lg"
                    className="w-full rounded-full bg-amber-500 hover:bg-amber-600 text-white font-bold shadow-md"
                    disabled={isItemAdd ? false : true}
                >
                    Order Now
                </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>Place this order?</AlertDialogTitle>
                    <AlertDialogDescription>
                        You won't be able to change your items or address after proceed.
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                        onClick={() => handleOrderPlace()}
                    >Continue</AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
}
